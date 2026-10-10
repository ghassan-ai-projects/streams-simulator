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
	if _, err := r.Advance(context.Background(), start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(pond, faultID, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if failureMode != "ok" {
		r.SetFailureMode(failureMode)
	}
	// Sealed label via the solver.
	solver, err := truth.NewSolver(spec, 11, 60*1e9, 24*3600*1e9)
	if err != nil {
		t.Fatal(err)
	}
	var setup []model.SetupCall
	if faultID == "aerator_failure" {
		setup = []model.SetupCall{{
			Effector: "start_aerator", EntityID: pond, CommandID: "setup",
			Args: map[string]any{"pond_id": pond, "level": 1.0}, AtNS: start,
		}}
	}
	gt, err := truth.BuildRecord(spec, solver, truth.Injection{
		ScenarioID: "aquaculture-pond/9001", Seed: 11, EntityID: pond, FaultID: faultID,
		OnsetNS: start + 2*3600*1e9, StartNS: start, EntityIDs: pondIDs, Setup: setup,
	})
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

// TestLoopResolvesWhenFaultOnsetLandsOnEmissionBoundary: the MCP harness
// pattern is inject-then-advance, so when the injected onset aligns exactly
// with an emission the sample at the onset already carries the fault. The
// pre-onset baseline must be the latest sample strictly before the onset,
// or the deviation collapses to zero and a correct recovery is scored as
// unresolved. Regression for the operator-endpoint golden loop.
func TestLoopResolvesWhenFaultOnsetLandsOnEmissionBoundary(t *testing.T) {
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
	if _, err := r.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	// Inject-then-advance, with the onset on an emission boundary.
	onset := start + 2*3600*1e9 // 06:00:00, an emission boundary
	if _, err := r.InjectFault(pond, "aerator_failure", onset, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+5*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	// Actuate and let the effect propagate.
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-boundary", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), r.World.Clock()+4*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(start), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-boundary", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	gt := &model.GroundTruthRecord{
		ScenarioID: "score/9002", Domain: "aquaculture-pond", Label: "aerator_failure",
		EntityID: pond, ExpectedEffector: "start_aerator", ExpectedEpisode: true,
		InjectionTimeNS: onset, FirstObservableTimeNS: onset, UnavoidableTimeNS: start + 3*3600*1e9,
		TrivialBaselineVerdict: model.TrivialNonTrivial,
	}
	sc, err := Score(r, gt)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.Loop.Resolved {
		t.Fatalf("recovery must resolve even when the onset lands on an emission boundary: %+v", sc.Loop)
	}
}
