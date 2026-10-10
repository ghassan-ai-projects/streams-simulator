package truth_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

func shippedSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestNewSolverRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := truth.NewSolver(nil, 1, 60e9, 3600e9); !errors.Is(err, truth.ErrNoSpec) {
		t.Fatalf("err = %v, want ErrNoSpec", err)
	}
}

func TestBuildRecordLabelsAPositiveAndANegativeClassScenario(t *testing.T) {
	t.Parallel()
	spec := shippedSpec(t)
	solver, err := truth.NewSolver(spec, 3, 60e9, 24*3600e9)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600e9
	ids := []string{"site-a/pond-1", "site-a/pond-2"}
	for _, tc := range []struct {
		fault    string
		positive bool
	}{{"aerator_failure", true}, {"transient_none", false}} {
		rec, err := truth.BuildRecord(spec, solver, truth.Injection{
			ScenarioID: "aquaculture-pond/" + tc.fault, Seed: 3, EntityID: ids[0], FaultID: tc.fault,
			OnsetNS: start + 2*3600e9, StartNS: start, EntityIDs: ids,
		})
		if err != nil {
			t.Fatal(err)
		}
		if rec.Label != tc.fault || rec.ExpectedEpisode != tc.positive || rec.IsNegativeClass == tc.positive {
			t.Fatalf("%s: label %+v", tc.fault, rec)
		}
	}
}

func TestStoreKeepsALabelSealedUntilItsRunCloses(t *testing.T) {
	t.Parallel()
	open := true
	store := truth.NewStore(func(string) bool { return open })
	if err := store.Seal("r-1", &model.GroundTruthRecord{Label: "f"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Reveal("r-1", false); err == nil {
		t.Fatal("an open run must not reveal its label")
	}
	open = false
	got, err := store.Reveal("r-1", false)
	if err != nil || got.Label != "f" {
		t.Fatalf("closed run: %v, %+v", err, got)
	}
	if sealed, unblinded, err := store.SealStatus("r-1"); err != nil || !sealed || unblinded {
		t.Fatalf("status = %v %v %v", sealed, unblinded, err)
	}
}
