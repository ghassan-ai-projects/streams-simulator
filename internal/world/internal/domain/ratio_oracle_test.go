package domain

import (
	"fmt"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

const ratioFloor = 1e-9

type sharing struct {
	weights, enabled []float64
	demand           float64
}

func input(state string) model.F1Input { return model.F1Input{State: state, Coef: 1, CoefSet: true} }

func (s sharing) world(t *testing.T) *World {
	t.Helper()
	return constWorld(t, func(spec *model.DomainSpec) {
		kind := &spec.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "demand", Initial: s.demand}, model.State{Name: "wsum"})
		terms := []model.F1Input{}
		for i := range s.weights {
			w, e, we := fmt.Sprintf("w%d", i), fmt.Sprintf("e%d", i), fmt.Sprintf("we%d", i)
			kind.State = append(kind.State, model.State{Name: w, Initial: s.weights[i]}, model.State{Name: e, Initial: s.enabled[i]},
				model.State{Name: we}, model.State{Name: fmt.Sprintf("share%d", i)})
			kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: we, Tier: "F1", F1: &model.F1Dyn{Form: model.FormProduct, Inputs: []model.F1Input{input(w), input(e)}}},
				model.Dynamics{Target: fmt.Sprintf("share%d", i), Tier: "F1", F1: &model.F1Dyn{Form: model.FormRatio, Floor: ratioFloor,
					Inputs: []model.F1Input{input(w), input(e), input("demand"), {State: "wsum", Coef: 1, CoefSet: true, Denominator: true}}}})
			terms = append(terms, input(we))
		}
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "wsum", Tier: "F1", F1: &model.F1Dyn{Form: model.FormSum, Inputs: terms}})
	}, model.DefaultStartTimeNS)
}

func (s sharing) closedForm(i int) float64 {
	total := 0.0
	for j, w := range s.weights {
		total += w * s.enabled[j]
	}
	if total == 0 {
		return 0
	}
	return s.weights[i] * s.enabled[i] * s.demand / total
}

func randomSharing(r *randutil.SplitMix64) sharing {
	n := 1 + r.Intn(5)
	s := sharing{weights: make([]float64, n), enabled: make([]float64, n), demand: 100 + 3000*r.Float64()}
	for i := range n {
		s.weights[i], s.enabled[i] = 0.5+r.Float64(), float64(r.Intn(2))
	}
	return s
}

func TestARatioSharesTheDemandByWeightAmongTheEnabledAndConservesIt(t *testing.T) {
	t.Parallel()
	r := randutil.NewSplitMix64(7)
	for c := range 60 {
		s := randomSharing(r)
		w := s.world(t)
		at := model.DefaultStartTimeNS + 60*secondsPerNS
		total, anyEnabled := 0.0, false
		for i := range s.weights {
			got, want := w.StateValue("e-1", fmt.Sprintf("share%d", i), at), s.closedForm(i)
			if math.Abs(got-want) > 1e-12*s.demand {
				t.Fatalf("case %d %+v: share%d = %v, closed form %v", c, s, i, got, want)
			}
			total += got
			anyEnabled = anyEnabled || s.enabled[i] == 1
		}
		if anyEnabled && math.Abs(total-s.demand) > 1e-12*s.demand {
			t.Fatalf("case %d %+v: shares add to %v, want the demand %v", c, s, total, s.demand)
		}
		if !anyEnabled && total != 0 {
			t.Fatalf("case %d %+v: no drive enabled, shares add to %v, want 0 (the floor divides a zero numerator)", c, s, total)
		}
	}
}

func TestARatioDividesByItsFloorWhenTheDenominatorIsAtOrBelowIt(t *testing.T) {
	t.Parallel()
	for _, den := range []float64{0, -2, 0.25, 0.5, 4} {
		w := constWorld(t, func(spec *model.DomainSpec) {
			kind := &spec.Kinds[0]
			kind.State = append(kind.State, model.State{Name: "num", Initial: 3}, model.State{Name: "den", Initial: den}, model.State{Name: "r"})
			kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "r", Tier: "F1", F1: &model.F1Dyn{Form: model.FormRatio, Floor: 0.5,
				Inputs: []model.F1Input{{State: "num", Coef: 2, CoefSet: true}, {State: "den", Coef: 1, CoefSet: true, Denominator: true}}}})
		}, model.DefaultStartTimeNS)
		if got, want := w.StateValue("e-1", "r", model.DefaultStartTimeNS+secondsPerNS), 6/math.Max(den, 0.5); got != want {
			t.Fatalf("2*3 / max(%v, 0.5) = %v, want %v", den, got, want)
		}
	}
}

func TestEqualWeightsGiveAFactorOfExactlyOne(t *testing.T) {
	t.Parallel()
	for n := 1; n <= 6; n++ {
		s := sharing{weights: make([]float64, n), enabled: make([]float64, n)}
		for i := range n {
			s.weights[i], s.enabled[i] = 1, 1
		}
		s.demand = float64(n)
		w := s.world(t)
		for i := range n {
			if got := w.StateValue("e-1", fmt.Sprintf("share%d", i), model.DefaultStartTimeNS+secondsPerNS); got != 1 {
				t.Fatalf("%d equal drives: share%d = %v, want exactly 1 (n / n)", n, i, got)
			}
		}
	}
}
