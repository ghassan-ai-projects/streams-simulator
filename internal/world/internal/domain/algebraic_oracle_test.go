package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func algebraicWorld(t *testing.T) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		in := func(state string, coef float64) model.F1Input { return model.F1Input{State: state, Coef: coef, CoefSet: true} }
		kind.State = append(kind.State, model.State{Name: "s"}, model.State{Name: "p"}, model.State{Name: "c"}, model.State{Name: "g"}, model.State{Name: "chain"})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "s", Tier: "F1", F1: &model.F1Dyn{Form: model.FormSum, Inputs: []model.F1Input{in("u", 1), in("x", -0.5)}}},
			model.Dynamics{Target: "p", Tier: "F1", F1: &model.F1Dyn{Form: model.FormProduct, Inputs: []model.F1Input{in("u", 3), in("x", 1)}, Clamp: &model.Clamp{Max: ptr(5)}}},
			model.Dynamics{Target: "c", Tier: "F1", F1: &model.F1Dyn{Form: model.FormCurve, Inputs: []model.F1Input{in("x", 1)},
				Table: &model.Table{X: []float64{0, 0.5, 1, 2}, Values: []float64{10, 20, 40, 50}}}},
			model.Dynamics{Target: "g", Tier: "F1", F1: &model.F1Dyn{Form: model.FormCurve, Inputs: []model.F1Input{in("x", 1), in("u", 1)},
				Table: &model.Table{X: []float64{0, 1}, Y: []float64{1, 3}, Grid: [][]float64{{0, 10}, {100, 300}}}}},
			model.Dynamics{Target: "chain", Tier: "F1", F1: &model.F1Dyn{Form: model.FormCurve, Inputs: []model.F1Input{in("s", 1)},
				Table: &model.Table{X: []float64{1, 2}, Values: []float64{0, 1}}}},
		)
	}, model.DefaultStartTimeNS)
}

func oneDimensional(x float64) float64 {
	switch {
	case x <= 0:
		return 10
	case x <= 0.5:
		return 10 + 20*x
	case x <= 1:
		return 20 + 40*(x-0.5)
	case x <= 2:
		return 40 + 10*(x-1)
	}
	return 50
}

func TestAlgebraicFormsAreExactFunctionsOfTheirInputsAtTheReadTime(t *testing.T) {
	t.Parallel()
	w := algebraicWorld(t)
	for _, seconds := range []float64{0, 0.4, 600, 1234.5, 3600, 20000} {
		at := model.DefaultStartTimeNS + int64(seconds*secondsPerNS)
		x, u := w.StateValue("e-1", "x", at), 2.0
		sum := u - 0.5*x
		checks := map[string]float64{
			"s":     sum,
			"p":     math.Min(3*u*x, 5),
			"c":     oneDimensional(x),
			"g":     (1-math.Min(x, 1))*(0+10*(u-1)/2) + math.Min(x, 1)*(100+200*(u-1)/2),
			"chain": math.Min(math.Max(sum, 1), 2) - 1,
		}
		for state, want := range checks {
			if got := w.StateValue("e-1", state, at); math.Abs(got-want) > 1e-12 {
				t.Fatalf("%s at %v s = %v, want %v (x = %v)", state, seconds, got, want, x)
			}
		}
	}
}

func TestAFaultOnAnAlgebraicStateOverlaysItsValue(t *testing.T) {
	t.Parallel()
	w := constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "s"})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "s", Tier: "F1", F1: &model.F1Dyn{Form: model.FormSum, Inputs: []model.F1Input{{State: "u", Coef: 2, CoefSet: true}}}})
		kind.Faults[0].Affects = []model.FaultEffect{{State: "s", Delta: 1}}
		kind.Faults[0].Onset.Magnitude = 1
	}, model.DefaultStartTimeNS)
	if _, err := w.InjectFault("e-1", "f1", model.DefaultStartTimeNS+10e9, nil); err != nil {
		t.Fatal(err)
	}
	if before, after := w.StateValue("e-1", "s", model.DefaultStartTimeNS+5e9), w.StateValue("e-1", "s", model.DefaultStartTimeNS+20e9); before != 4 || after != 5 {
		t.Fatalf("sum before and after a +1 fault = %v, %v; want 4, 5", before, after)
	}
}

func TestACurveClampsAtItsEndsAndHitsItsPoints(t *testing.T) {
	t.Parallel()
	table := &model.Table{X: []float64{0, 0.5, 1, 2}, Values: []float64{10, 20, 40, 50}}
	for _, x := range []float64{-3, 0, 0.25, 0.5, 0.75, 1, 1.5, 2, 9} {
		if got, want := curveValue(table, []float64{x}), oneDimensional(x); math.Abs(got-want) > 1e-12 {
			t.Fatalf("curve(%v) = %v, want %v", x, got, want)
		}
	}
}
