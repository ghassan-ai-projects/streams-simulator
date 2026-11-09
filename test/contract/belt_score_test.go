package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const wireKey = "sha256:c16de560cda622f0c59403923fd0f84feb94315f91020c1e5fa3662e5e104114"

func writeRunFile(t *testing.T, dir, name string, document any) {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

// limpFeedRun is plant B with D-3's inverter tripping at 60 s and a plant-wire host cutting the feed at 70 s: the world
// logs the call under the DCMD's idempotency key, the host decided it at 65 s, before the node applied it.
func limpFeedRun(t *testing.T, dir string) string {
	t.Helper()
	spec := beltPlant(t, "b")
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 7, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("D-3", "conveyor_drive__vfd_fault", beltNS(60), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("bulk_feeder__set_feed", "F-1", wireKey, map[string]any{"percent": 18.0}, beltNS(70)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), beltNS(300), false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return r.ID
}

func vfdLabel(t *testing.T) *model.GroundTruthRecord {
	t.Helper()
	spec := beltPlant(t, "b")
	solver, err := truth.NewSolver(spec, 7, 120*1e9, 0)
	if err != nil {
		t.Fatal(err)
	}
	label, err := truth.BuildRecord(spec, solver, truth.Injection{ScenarioID: "belt-conveyor-drives-b/0001", Seed: 7, EntityID: "D-3",
		FaultID: "conveyor_drive__vfd_fault", OnsetNS: beltNS(60), StartNS: model.DefaultStartTimeNS, EntityIDs: world.PopulationIDs(spec, 0)})
	if err != nil {
		t.Fatal(err)
	}
	return label
}

func TestAPlantWireVerdictIsCreditedForTheFeedCutItDecidedBeforeTheWorldAppliedIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runID := limpFeedRun(t, dir)
	label := vfdLabel(t)
	if want := []model.ExpectedAction{{EntityID: "F-1", Effector: "bulk_feeder__set_feed"}}; !slices.Equal(label.ExpectedActions, want) {
		t.Fatalf("label expects %v, want %v", label.ExpectedActions, want)
	}
	writeRunFile(t, dir, model.FileLabel, label)
	writeRunFile(t, dir, model.FileVerdict, model.Verdict{SchemaVersion: "0.1", RunID: runID, Consumer: model.ConsumerInfo{Name: "host", Version: "1"},
		Detections: []model.Detection{{EntityID: "D-3", DetectedAt: model.FormatTime(beltNS(65)), Label: "conveyor_drive__vfd_fault"}},
		Actions: []model.Action{{CommandID: wireKey, Effector: "bulk_feeder__set_feed", EntityID: "F-1", IssuedAt: model.FormatTime(beltNS(65)),
			OutcomeBelieved: model.BelievedSucceeded}}})
	graded, err := grading.GradeRun(t.Context(), grading.Options{Artifact: filepath.Join(dir, model.FileRunArtifact)})
	if err != nil {
		t.Fatal(err)
	}
	card := graded.Scorecards[0]
	if !card.Consumer.ActionFidelity || !card.Loop.ActionAppropriate || !card.Judgment.Detected || !card.Judgment.LabelCorrect {
		t.Fatalf("consumer %+v loop %+v judgment %+v, want the feed cut faithful and appropriate and D-3 detected", card.Consumer, card.Loop, card.Judgment)
	}
	if len(graded.Unexpected.Actions) != 0 || graded.Unexpected.EffectCalls != 0 {
		t.Fatalf("unexpected = %+v, want the F-1 call graded with D-3's fault", graded.Unexpected)
	}
}
