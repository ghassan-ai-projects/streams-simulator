package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type tankPair struct{ upper, lower float64 }

func transferLevels(from tankPair, transfer, seconds float64) tankPair {
	const feed, outlet, capUpper, capLower = 0.02, 0.015, 4.0, 2.0
	a := [2][2]float64{{-transfer / capUpper, transfer / capUpper}, {transfer / capLower, -(transfer + outlet) / capLower}}
	det := a[0][0]*a[1][1] - a[0][1]*a[1][0]
	steady := tankPair{lower: feed / outlet}
	steady.upper = steady.lower + feed/transfer
	half := (a[0][0] + a[1][1]) / 2
	q := math.Sqrt(half*half - det)
	c, s := math.Cosh(q*seconds), math.Sinh(q*seconds)/q
	du, dl := from.upper-steady.upper, from.lower-steady.lower
	decay := math.Exp(half * seconds)
	return tankPair{
		upper: steady.upper + decay*((c+s*(a[0][0]-half))*du+s*a[0][1]*dl),
		lower: steady.lower + decay*(s*a[1][0]*du+(c+s*(a[1][1]-half))*dl),
	}
}

func TestTheSolverFindsTheClosedFormCrossingOfAPipeBlockage(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.TransferDomain())
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	const onsetS, sample, sigma = 1800.0, 60 * 1e9, 0.01
	onset := start + int64(onsetS*1e9)
	atOnset := transferLevels(tankPair{upper: 2, lower: 1}, 0.01, onsetS)
	crossing := func(snr float64) int64 {
		for at := onset; ; at += sample {
			elapsed := float64(at-onset) / 1e9
			gap := transferLevels(atOnset, 0.01, elapsed).lower - transferLevels(atOnset, 0.004, elapsed).lower
			if gap >= snr*sigma {
				return at
			}
		}
	}
	res, err := NewSolver(spec, 42, sample, 6*3600*1e9).solve("cell-1", "pipe_blockage", onset, start, []string{"cell-1", "cell-2", "cell-3"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Observable || res.FirstObservableNS != crossing(3) || res.UnavoidableNS != crossing(10) {
		t.Fatalf("first %d unavoidable %d, closed form %d and %d", res.FirstObservableNS, res.UnavoidableNS, crossing(3), crossing(10))
	}
}

func TestTheTransferWorldFollowsTheMatrixExponential(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.TransferDomain())
	if err != nil {
		t.Fatal(err)
	}
	solver := NewSolver(spec, 42, 0, 0)
	w, err := solver.newOracleWorld(model.DefaultStartTimeNS, []string{"cell-1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, seconds := range []float64{60, 1800, 7200, 4 * 3600} {
		want := transferLevels(tankPair{upper: 2, lower: 1}, 0.01, seconds)
		at := model.DefaultStartTimeNS + int64(seconds*1e9)
		upper, lower := w.StateValue("cell-1", "upper", at), w.StateValue("cell-1", "lower", at)
		if math.Abs(upper-want.upper) > 1e-6 || math.Abs(lower-want.lower) > 1e-6 {
			t.Fatalf("t=%v: got %v %v, matrix exponential %v %v", seconds, upper, lower, want.upper, want.lower)
		}
	}
}
