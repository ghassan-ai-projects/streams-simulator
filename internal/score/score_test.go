package score

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

const (
	aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"
	nativeAdapter   = "../../adapters/native-jsonl.adapter.json"
)

var pondIDs = []string{"site-a/pond-1", "site-a/pond-2", "site-a/pond-3", "site-a/pond-4",
	"site-a/pond-5", "site-a/pond-6", "site-a/pond-7", "site-a/pond-8"}

func testBase(t *testing.T) (*domain.Compiled, *model.Adapter) {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	return spec, a
}

// setupFaultedRun builds a run with a pre-injected fault and the aerator
// running (so aerator_failure is real). The failure mode is set after setup
// so the scenario context applies deterministically.
func setupFaultedRun(t *testing.T, failureMode, faultID string) (*run.Run, *model.GroundTruthRecord) {
	t.Helper()
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
		ForceFailureMode: "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	pond := "site-a/pond-1"
	// Night aerator run.
	if _, err := r.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(pond, faultID, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if failureMode != "ok" {
		r.SetFailureMode(failureMode)
	}
	// Sealed label via the solver.
	solver := truth.NewSolver(spec, 11, 60*1e9, 24*3600*1e9)
	var setup []truth.SetupCall
	if faultID == "aerator_failure" {
		setup = []truth.SetupCall{{
			Effector: "start_aerator", EntityID: pond, CommandID: "setup",
			Args: map[string]any{"pond_id": pond, "level": 1.0}, AtNS: start,
		}}
	}
	gt, err := truth.BuildRecord(spec, solver, "aquaculture-pond/9001", 11,
		pond, faultID, start+2*3600*1e9, start, pondIDs, false, nil, setup)
	if err != nil {
		t.Fatal(err)
	}
	return r, gt
}

func submitVerdict(t *testing.T, r *run.Run, actions []model.Action, detections []model.Detection) {
	t.Helper()
	v := &model.Verdict{
		SchemaVersion: "0.1",
		RunID:         r.ID,
		Consumer:      model.ConsumerInfo{Name: "test", Version: "0.1"},
		Detections:    detections,
		Actions:       actions,
	}
	if err := r.SubmitVerdict(v); err != nil {
		t.Fatal(err)
	}
}

// loopResolved is a test helper: did the aerator output recover?
func loopResolved(r *run.Run) bool {
	return r.World.StateValue("site-a/pond-1", "aerator_output", r.World.Clock()) > 0.9
}

// TestSilentNoEffectFalseSuccess is the highest-value single test in the
// plan: under silent_no_effect the confirmation channel reports the
// counterfactual (the motor draws current), so the evidence is internally
// consistent with success — only the absent physical outcome betrays the
// failure. A consumer reporting success has recorded a false outcome.
func TestSilentNoEffectFalseSuccess(t *testing.T) {
	r, gt := setupFaultedRun(t, "silent_no_effect", "aerator_failure")
	pond := "site-a/pond-1"
	// Actuate; the shadow state takes the effect, the real world does not.
	res, err := r.InvokeEffector("start_aerator", pond, "cmd-91", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Accepted {
		t.Fatalf("silent_no_effect must ack success: %+v", res)
	}
	// The effect propagates.
	if _, err := r.Advance(r.World.Clock()+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	at := r.World.Clock()
	// The confirmation channel reports the counterfactual: current drawn.
	current := r.World.Reading(pond, "pond.aerator_current", at)
	if current < 10.0 { // gain 18 x shadow aerator_output 1.0 + offset 0.4
		t.Fatalf("silent_no_effect must shadow the confirmation channel: current=%v", current)
	}
	// But the physical outcome never arrives: DO stays at the faulted level.
	do := r.World.StateValue(pond, "dissolved_oxygen_true", at)
	if do > 6.0 {
		t.Fatalf("real DO must not recover under silent_no_effect: %v", do)
	}
	// A dishonest consumer reports success.
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(at), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-91", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(at), OutcomeBelieved: model.BelievedSucceeded},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.Loop.FalseSuccess || sc.Loop.FalseSuccessRate != 1.0 {
		t.Fatalf("false success must be caught: %+v", sc.Loop)
	}
}

// TestSilentNoEffectHonestConsumer: the same scenario with an honest
// consumer reports unknown and the metric reads zero.
func TestSilentNoEffectHonestConsumer(t *testing.T) {
	r, gt := setupFaultedRun(t, "silent_no_effect", "aerator_failure")
	pond := "site-a/pond-1"
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-92", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(r.World.Clock()+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-92", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedUnknown},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Loop.FalseSuccess || sc.Loop.FalseSuccessRate != 0.0 {
		t.Fatalf("honest consumer must not be flagged: %+v", sc.Loop)
	}
}

// TestClosedLoopRecoveryAndIdempotency: with mode ok the effector recovers
// the state and the loop metrics credit it.
func TestClosedLoopRecoveryAndIdempotency(t *testing.T) {
	r, gt := setupFaultedRun(t, "ok", "aerator_failure")
	pond := "site-a/pond-1"
	res, err := r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Accepted {
		t.Fatal("ok mode must accept")
	}
	if _, err := r.Advance(r.World.Clock()+4*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if !loopResolved(r) {
		t.Fatal("effect must resolve the fault")
	}
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-1", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.Loop.Resolved || sc.Loop.FalseSuccess {
		t.Fatalf("loop metrics wrong: %+v", sc.Loop)
	}
	if !sc.Instrument.EffectorIdempotency {
		t.Fatalf("idempotency flag wrong: %+v", sc.Instrument)
	}
	if !sc.Consumer.ActionFidelity {
		t.Fatalf("action fidelity wrong: %+v", sc.Consumer)
	}
}

// TestSuspiciousEarlyDetection: a detection before first_observable_time is
// flagged, never credited. The fouling fault has a ~2.6h observability lag,
// so a detection right after injection is before the noise floor.
func TestSuspiciousEarlyDetection(t *testing.T) {
	r, gt := setupFaultedRun(t, "ok", "do_probe_fouling")
	pond := "site-a/pond-1"
	early := model.FormatTime(gt.InjectionTimeNS + 10*1e9)
	submitVerdict(t, r, nil, []model.Detection{{
		EntityID: pond, DetectedAt: early, Label: "do_probe_fouling",
	}})
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.Judgment.Suspicious || sc.Judgment.Detected {
		t.Fatalf("early detection must be suspicious, not credited: %+v", sc.Judgment)
	}
}

func TestJudgmentChoosesEarliestAndRequiresLabel(t *testing.T) {
	r, gt := setupFaultedRun(t, "ok", "do_probe_fouling")
	pond := "site-a/pond-1"
	later := model.FormatTime(gt.FirstObservableTimeNS + 20*60*1e9)
	earlier := model.FormatTime(gt.FirstObservableTimeNS + 10*60*1e9)
	submitVerdict(t, r, nil, []model.Detection{
		{EntityID: pond, DetectedAt: later, Label: gt.Label},
		{EntityID: pond, DetectedAt: earlier},
	})
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Judgment.DetectionLatencyNS != 10*60*1e9 || sc.Judgment.LabelCorrect {
		t.Fatalf("judgment must use earliest detection and require exact label: %+v", sc.Judgment)
	}
}

func TestNegativeUnobservableDetectionIsFalsePositive(t *testing.T) {
	spec, a := testBase(t)
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 21, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS,
	})
	if err != nil {
		t.Fatal(err)
	}
	gt := &model.GroundTruthRecord{ScenarioID: "aquaculture-pond/9999", Domain: spec.Spec.ID,
		EntityID: pondIDs[0], Label: "negative", IsNegativeClass: true, ExpectedEpisode: false}
	submitVerdict(t, r, nil, []model.Detection{{EntityID: pondIDs[0], DetectedAt: model.FormatTime(model.DefaultStartTimeNS)}})
	m := judgment(r, gt)
	if !m.FalsePositive {
		t.Fatalf("negative detection must remain a false positive even without observable time: %+v", m)
	}
}

func TestDroppedDetectionRequiresOneDetectionPerDrop(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 22, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 1.0}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if err := r.SubmitVerdict(&model.Verdict{
		SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "test", Version: "1"},
		Detections: []model.Detection{{EntityID: pondIDs[0], DetectedAt: model.FormatTime(start + 1*1e9)}},
	}); err != nil {
		t.Fatal(err)
	}
	m := consumer(r, &model.GroundTruthRecord{EntityID: pondIDs[0]})
	if m.DroppedEventDetection {
		t.Fatal("one detection must not satisfy every dropped delivery")
	}
}

func TestActionFidelityChecksEffectorAndEntity(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 23, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, ForceFailureMode: "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", pondIDs[0], "cmd", map[string]any{"pond_id": pondIDs[0], "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if err := r.SubmitVerdict(&model.Verdict{
		SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "test", Version: "1"},
		Actions: []model.Action{{CommandID: "cmd", Effector: "halt_feeding", EntityID: pondIDs[1], IssuedAt: model.FormatTime(start), OutcomeBelieved: model.BelievedSucceeded}},
	}); err != nil {
		t.Fatal(err)
	}
	if consumer(r, &model.GroundTruthRecord{}).ActionFidelity {
		t.Fatal("wrong effector/entity must not receive action credit")
	}
}
