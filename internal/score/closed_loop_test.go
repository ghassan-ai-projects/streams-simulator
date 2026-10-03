package score

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

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
	if _, err := r.Advance(context.Background(), r.World.Clock()+4*3600*1e9, false); err != nil {
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
