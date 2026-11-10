package contract

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	vibrationFrozen = "conveyor_drive__vibration__frozen"
	flatnessSeenS   = 900
)

// sessionLabelOnD3 builds a disturbance's label on D-3 the way a plant-wire session of plant A writes labels.json.
func sessionLabelOnD3(t *testing.T, seed uint64, fault string, onset int64) *model.GroundTruthRecord {
	t.Helper()
	spec := beltPlant(t, "a")
	solver, err := truth.NewSolver(spec, seed, 120*1e9, 0)
	if err != nil {
		t.Fatal(err)
	}
	label, err := truth.BuildRecord(spec, solver, truth.Injection{ScenarioID: "belt-conveyor-drives-a/0001", Seed: seed, EntityID: "D-3",
		FaultID: fault, OnsetNS: onset, StartNS: model.DefaultStartTimeNS, EntityIDs: world.PopulationIDs(spec, 0)})
	if err != nil {
		t.Fatal(err)
	}
	return label
}

// A frozen reading of a noisy channel is flat from its injection (O26), so a consumer that reports it from that
// flatness 15 minutes later is graded detected, not too early.
func TestTheFrozenVibrationOfTheBearingStoryIsObservableAtItsInjectionAndItsReportIsGradedDetected(t *testing.T) {
	t.Parallel()
	scenario := committedScenario(t, "s5-a.do-nothing.scenario.json")
	_, onset := injectionOnD3(t, scenario, vibrationFrozen)
	label := sessionLabelOnD3(t, scenario.Seed, vibrationFrozen, onset)
	if !label.IsPositive() || label.FirstObservableTimeNS != onset || len(label.ExpectedActions) != 0 {
		t.Fatalf("frozen label: positive %v, first observable %d (injected %d), expected actions %v; want positive, observable at its injection, no action",
			label.IsPositive(), label.FirstObservableTimeNS, onset, label.ExpectedActions)
	}
	reportedAt := onset + flatnessSeenS*1e9
	verdict := &model.Verdict{SchemaVersion: "0.1", RunID: "r-s5-a",
		Detections: []model.Detection{{EntityID: "D-3", DetectedAt: model.FormatTime(reportedAt), Label: vibrationFrozen}}}
	graded, err := score.ScoreRun(score.Evidence{RunID: "r-s5-a", Verdict: verdict, Ledger: []model.LedgerRecord{}}, []*model.GroundTruthRecord{label})
	if err != nil {
		t.Fatal(err)
	}
	if j := graded.Scorecards[0].Judgment; !j.Detected || j.Suspicious || !j.LabelCorrect || j.DetectionLatencyNS != flatnessSeenS*1e9 {
		t.Fatalf("judgment %+v, want detected with the right label %d s after the freeze, not suspicious", j, flatnessSeenS)
	}
}
