package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheSolverFindsTheClosedFormCrossingOfAFaultAddedByAnOverride(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"tank-1", "tank-2", "T-main", "pump-1", "pump-2", "pump-3", "P-slow", "air_header"}
	start := model.DefaultStartTimeNS
	onset := start + 3600*1e9
	const sample = 60 * 1e9
	solver := NewSolver(spec, 42, sample, 6*3600*1e9)
	const slopePerHour, sigma = 5.0, 0.1
	res, err := solver.solve("T-main", "tank_overheat", onset, start, ids, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantFirst := firstGridTime(onset, sample, 3.0*sigma/slopePerHour*3600)
	wantUnavoidable := firstGridTime(onset, sample, 10.0*sigma/slopePerHour*3600)
	if !res.Observable || res.FirstObservableNS != wantFirst || res.UnavoidableNS != wantUnavoidable {
		t.Fatalf("first %d unavoidable %d, closed form %d and %d", res.FirstObservableNS, res.UnavoidableNS, wantFirst, wantUnavoidable)
	}
	if math.Abs(res.EffectiveSigma-sigma) > 1e-12 {
		t.Fatalf("effective sigma %v, want %v", res.EffectiveSigma, sigma)
	}
}
