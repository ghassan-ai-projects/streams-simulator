package truth

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"

func loadSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func entityIDs(spec *domain.Compiled) []string {
	n := spec.Spec.Entities.Count.Default
	var ids []string
	for i := 1; i <= n; i++ {
		ids = append(ids, "site-a/pond-"+string(rune('0'+i)))
	}
	return ids
}

func TestAeratorFailureImmediatelyObservable(t *testing.T) {
	spec := loadSpec(t)
	ids := entityIDs(spec)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	onset := start + 2*3600*1e9
	// Scenario context: the aerator runs through the night.
	setup := []SetupCall{{
		Effector: "start_aerator", EntityID: "site-a/pond-1", CommandID: "setup",
		Args: map[string]any{"pond_id": "site-a/pond-1", "level": 1.0}, AtNS: start,
	}}
	solver := NewSolver(spec, 42, 60*1e9, 24*3600*1e9)
	res, err := solver.Solve("site-a/pond-1", "aerator_failure", onset, start, ids, setup)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Observable {
		t.Fatal("aerator_failure must be observable when the aerator is running")
	}
	// The aerator current channel drops by 18 A (gain 18 x delta -1.0)
	// against sigma 0.2: far past SNR 3 immediately.
	if res.FirstObservableNS != onset && math.Abs(float64(res.FirstObservableNS-onset)) > 60*1e9 {
		t.Fatalf("first observable should be at onset, got %d vs %d", res.FirstObservableNS, onset)
	}
	if res.FirstObservableNS > res.UnavoidableNS {
		t.Fatalf("first observable after unavoidable: %d > %d", res.FirstObservableNS, res.UnavoidableNS)
	}
	if math.Abs(res.EffectiveSigma-0.2) > 1e-9 {
		t.Fatalf("single-channel sigma wrong: %v", res.EffectiveSigma)
	}
}

func TestProbeFoulingObservabilityLag(t *testing.T) {
	// The lethal sensor fault: the probe reads high and stable, so the
	// single-channel SNR never crosses; only the peer residual against
	// sibling ponds makes it observable, and only after the fouling ramp
	// accumulates. first_observable must be hours after injection.
	spec := loadSpec(t)
	ids := entityIDs(spec)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	onset := start + 2*3600*1e9
	solver := NewSolver(spec, 7, 60*1e9, 24*3600*1e9)
	res, err := solver.Solve("site-a/pond-1", "do_probe_fouling", onset, start, ids, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Observable {
		t.Fatal("fouling must become observable within the horizon")
	}
	lagHours := float64(res.FirstObservableNS-onset) / 3600e9
	if lagHours < 1.5 || lagHours > 5 {
		t.Fatalf("first observable should lag ~2.6h (fouling 0.04/h, coef 2.5, sigma 0.086): got %.2fh", lagHours)
	}
	// The peer residual pools sibling variance.
	if res.EffectiveSigma <= 0.08 || res.EffectiveSigma > 0.1 {
		t.Fatalf("peer sigma should pool ~0.086, got %v", res.EffectiveSigma)
	}
}

func TestBuildRecord(t *testing.T) {
	spec := loadSpec(t)
	ids := entityIDs(spec)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	onset := start + 2*3600*1e9
	solver := NewSolver(spec, 3, 60*1e9, 24*3600*1e9)
	setup := []SetupCall{{
		Effector: "start_aerator", EntityID: "site-a/pond-1", CommandID: "setup",
		Args: map[string]any{"pond_id": "site-a/pond-1", "level": 1.0}, AtNS: start,
	}}
	rec, err := BuildRecord(spec, solver, "aquaculture-pond/0001", 3,
		"site-a/pond-1", "aerator_failure", onset, start, ids, false, []string{"drop@0.01"}, setup)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Label != "aerator_failure" || rec.ExpectedEpisode != true || rec.IsNegativeClass != false {
		t.Fatalf("label wrong: %+v", rec)
	}
	if rec.DeadlineNS != rec.FirstObservableTimeNS+5400*1e9 {
		t.Fatalf("deadline wrong: %d vs %d", rec.DeadlineNS, rec.FirstObservableTimeNS+5400*1e9)
	}
	if rec.ExpectedEffector != "start_aerator" {
		t.Fatalf("expected effector wrong: %s", rec.ExpectedEffector)
	}
	if len(rec.Perturbations) != 1 || rec.Perturbations[0] != "drop@0.01" {
		t.Fatalf("perturbations wrong: %v", rec.Perturbations)
	}

	// Negative class: expected_episode false.
	recN, err := BuildRecord(spec, solver, "aquaculture-pond/0002", 3,
		"site-a/pond-2", "transient_none", onset, start, ids, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if recN.IsNegativeClass != true || recN.ExpectedEpisode != false {
		t.Fatalf("negative class wrong: %+v", recN)
	}
}

func TestStoreSealing(t *testing.T) {
	s := NewStore()
	rec := &model.GroundTruthRecord{ScenarioID: "x/0001", Label: "f"}
	s.Seal("r-1", rec)
	// Open run: reveal refused without unblind.
	s.OpenChecker = func(runID string) bool { return runID == "r-1" }
	if _, err := s.Reveal("r-1", false); err == nil {
		t.Fatal("reveal on an open run must be refused")
	}
	got, err := s.Reveal("r-1", true)
	if err != nil {
		t.Fatalf("unblind reveal refused: %v", err)
	}
	if got.Label != "f" {
		t.Fatalf("wrong label: %+v", got)
	}
	if _, err := s.Reveal("r-nope", false); err == nil {
		t.Fatal("unknown run must fail")
	}
}
