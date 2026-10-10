package truth_test

import (
	"errors"
	"path/filepath"
	"strings"
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

func TestBuildRecordRefusesMissingInputsAndUnknownFaults(t *testing.T) {
	t.Parallel()
	spec := shippedSpec(t)
	solver, err := truth.NewSolver(spec, 3, 60e9, 24*3600e9)
	if err != nil {
		t.Fatal(err)
	}
	in := truth.Injection{ScenarioID: "s", EntityID: "site-a/pond-1", FaultID: "aerator_failure"}
	if _, err := truth.BuildRecord(nil, solver, in); !errors.Is(err, truth.ErrNoSpec) {
		t.Fatalf("nil spec: %v", err)
	}
	if _, err := truth.BuildRecord(spec, nil, in); !errors.Is(err, truth.ErrNoSolver) {
		t.Fatalf("nil solver: %v", err)
	}
	in.FaultID = "no_such_fault"
	if _, err := truth.BuildRecord(spec, solver, in); err == nil || !strings.Contains(err.Error(), `unknown fault "no_such_fault"`) {
		t.Fatalf("unknown fault: %v", err)
	}
}

func TestBuildRecordLabelsAScenarioFromTheSolvedOnsets(t *testing.T) {
	t.Parallel()
	spec := shippedSpec(t)
	solver, err := truth.NewSolver(spec, 3, 60e9, 24*3600e9)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600e9
	ids := []string{"site-a/pond-1", "site-a/pond-2"}
	perturbations := []string{"drop@0.01"}
	rec, err := truth.BuildRecord(spec, solver, truth.Injection{
		ScenarioID: "aquaculture-pond/0001", Seed: 3, EntityID: ids[0], FaultID: "aerator_failure",
		OnsetNS: start + 2*3600e9, StartNS: start, EntityIDs: ids, Perturbations: perturbations,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Label != "aerator_failure" || !rec.ExpectedEpisode || rec.InjectionTimeNS != start+2*3600e9 {
		t.Fatalf("label = %+v", rec)
	}
	perturbations[0] = "mutated"
	if rec.Perturbations[0] != "drop@0.01" {
		t.Fatal("the record must not share the caller's perturbation slice")
	}
}

func TestStoreUnblindStampsTheRunAndRefusesDoubleSealAndUnknownRuns(t *testing.T) {
	t.Parallel()
	store := truth.NewStore(func(string) bool { return true })
	if err := store.Seal("r-1", &model.GroundTruthRecord{Label: "f"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Seal("r-1", &model.GroundTruthRecord{Label: "g"}); err == nil || !strings.Contains(err.Error(), "already sealed") {
		t.Fatalf("second seal: %v", err)
	}
	if _, err := store.Reveal("r-1", false); err == nil {
		t.Fatal("an open run must not reveal without unblind")
	}
	if got, err := store.Reveal("r-1", true); err != nil || got.Label != "f" {
		t.Fatalf("unblinded reveal: %v, %+v", err, got)
	}
	if sealed, unblinded, err := store.SealStatus("r-1"); err != nil || !sealed || !unblinded {
		t.Fatalf("status after unblind = %v %v %v", sealed, unblinded, err)
	}
	if _, err := store.Reveal("r-2", true); err == nil || !strings.Contains(err.Error(), "no sealed label") {
		t.Fatalf("unknown run reveal: %v", err)
	}
	if _, _, err := store.SealStatus("r-2"); err == nil || !strings.Contains(err.Error(), "unknown run") {
		t.Fatalf("unknown run status: %v", err)
	}
}

func TestStoreWithoutAnOpenRunCheckOrBackingStoreFailsClosed(t *testing.T) {
	t.Parallel()
	nilCheck := truth.NewStore(nil)
	if err := nilCheck.Seal("r-1", &model.GroundTruthRecord{Label: "f"}); err != nil {
		t.Fatal(err)
	}
	if _, err := nilCheck.Reveal("r-1", false); err == nil {
		t.Fatal("a store without an open-run check must refuse to reveal")
	}
	var zero truth.Store
	if err := zero.Seal("r-1", &model.GroundTruthRecord{}); !errors.Is(err, truth.ErrNoStore) {
		t.Fatalf("zero store seal: %v", err)
	}
	if _, err := zero.Reveal("r-1", true); !errors.Is(err, truth.ErrNoStore) {
		t.Fatalf("zero store reveal: %v", err)
	}
}
