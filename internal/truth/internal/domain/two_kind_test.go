package domain

import (
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twoKindFixture(t *testing.T) (*domain.Compiled, []string) {
	t.Helper()
	spec, err := domain.Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	return spec, []string{"tank-1", "tank-2", "tank-3", "pump-1", "pump-2", "pump-3", "pump-4"}
}

func firstGridTime(onset, sampleNS int64, crossingS float64) int64 {
	steps := math.Ceil(crossingS * 1e9 / float64(sampleNS))
	return onset + int64(steps)*sampleNS
}

func TestTheSolverFindsTheClosedFormCrossingOfEachKind(t *testing.T) {
	t.Parallel()
	spec, ids := twoKindFixture(t)
	start := model.DefaultStartTimeNS
	onset := start + 3600*1e9
	const sample = 60 * 1e9
	solver := NewSolver(spec, 42, sample, 6*3600*1e9)
	peerSigma := 0.1 * math.Sqrt(1+1.0/3)
	cases := []struct {
		entity, fault        string
		rateDeltaPerHr       float64
		firstSNR, unavoidSNR float64
		wantSigma            float64
	}{
		{"tank-2", "tank_leak", 1.0, 3.2, 10.8, 0.05},
		{"pump-3", "pump_wear", 4.0, 3.0, 10.0, peerSigma},
	}
	for _, tc := range cases {
		res, err := solver.solve(tc.entity, tc.fault, onset, start, ids, nil)
		if err != nil {
			t.Fatalf("%s: %v", tc.fault, err)
		}
		wantFirst := firstGridTime(onset, sample, tc.firstSNR*tc.wantSigma/tc.rateDeltaPerHr*3600)
		wantUnavoidable := firstGridTime(onset, sample, tc.unavoidSNR*tc.wantSigma/tc.rateDeltaPerHr*3600)
		if !res.Observable || res.FirstObservableNS != wantFirst || res.UnavoidableNS != wantUnavoidable {
			t.Fatalf("%s: first %d unavoidable %d observable %v, closed form %d and %d", tc.fault, res.FirstObservableNS, res.UnavoidableNS, res.Observable, wantFirst, wantUnavoidable)
		}
		if math.Abs(res.EffectiveSigma-tc.wantSigma) > 1e-12 {
			t.Fatalf("%s: effective sigma %v, want %v", tc.fault, res.EffectiveSigma, tc.wantSigma)
		}
	}
}

func TestAPeerResidualPoolsOnlyTheEntitiesOfTheSameKind(t *testing.T) {
	t.Parallel()
	spec, ids := twoKindFixture(t)
	solver := NewSolver(spec, 42, 60*1e9, 6*3600*1e9)
	if got, want := solver.peerSigma("pump.flow", 4), 0.1*math.Sqrt(1+1.0/3); math.Abs(got-want) > 1e-12 {
		t.Fatalf("peerSigma = %v, want %v", got, want)
	}
	start := model.DefaultStartTimeNS
	clean, err := solver.buildWorld("pump-1", start, ids, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if peers := clean.PeerIDs("pump-1"); len(peers) != 4 {
		t.Fatalf("pump peers = %v, want the four pumps and no tank", peers)
	}
	if sum, count := peerReadings("pump-1", "pump.flow", clean, start+3600*1e9); count != 3 || sum <= 0 {
		t.Fatalf("peer readings = %v over %d, want three pump peers", sum, count)
	}
}

func TestTheSolverRefusesAFaultOfAnotherKind(t *testing.T) {
	t.Parallel()
	spec, ids := twoKindFixture(t)
	start := model.DefaultStartTimeNS
	solver := NewSolver(spec, 42, 60*1e9, 6*3600*1e9)
	if _, err := solver.solve("pump-1", "tank_leak", start+3600*1e9, start, ids, nil); err == nil || !strings.Contains(err.Error(), "is not declared by the kind") {
		t.Fatalf("solve = %v, want a refusal of a tank fault on a pump", err)
	}
}
