package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func chainWorld(t *testing.T, sourceForm string) *World {
	t.Helper()
	return constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "a", Initial: 0}, model.State{Name: "y", Initial: 0})
		source := &model.F1Dyn{Form: sourceForm, Capacity: 4, TimeConstantS: 120, Gain: 1, Inputs: []model.F1Input{{State: "u", Coef: 1, CoefSet: true}}}
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "a", Tier: "F1", DTMs: 1000, F1: source},
			model.Dynamics{Target: "y", Tier: "F1", DTMs: 5000, F1: &model.F1Dyn{Form: "first_order_lag", TimeConstantS: 60, Gain: 1, Inputs: []model.F1Input{{State: "a", Coef: 1, CoefSet: true}}}})
	}, model.DefaultStartTimeNS)
}

func TestAStateFedByAnIntegratedStateDoesNotDependOnReadOrder(t *testing.T) {
	t.Parallel()
	for _, form := range []string{model.FormStore, "first_order_lag"} {
		at := model.DefaultStartTimeNS + 600e9 + 300e6
		yFirst := chainWorld(t, form)
		y1, a1 := yFirst.StateValue("e-1", "y", at), yFirst.StateValue("e-1", "a", at)
		aFirst := chainWorld(t, form)
		a2, y2 := aFirst.StateValue("e-1", "a", at), aFirst.StateValue("e-1", "y", at)
		often := chainWorld(t, form)
		for read := model.DefaultStartTimeNS; read < at; read += 7e9 {
			often.StateValue("e-1", "a", read)
		}
		y3, a3 := often.StateValue("e-1", "y", at), often.StateValue("e-1", "a", at)
		if y1 != y2 || y1 != y3 || a1 != a2 || a1 != a3 {
			t.Fatalf("%s: read order changed the values: y %v %v %v, a %v %v %v", form, y1, y2, y3, a1, a2, a3)
		}
	}
}

func TestAStoreLevelStaysWithinItsStateBounds(t *testing.T) {
	t.Parallel()
	w := constWorld(t, func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "a", Initial: 9, Min: 0, Max: 10})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "a", Tier: "F1", DTMs: 1000,
			F1: &model.F1Dyn{Form: model.FormStore, Capacity: 1, Inputs: []model.F1Input{{State: "u", Coef: 1, CoefSet: true}}}})
		kind.Flows = []model.Flow{{Name: "outlet", From: "a", Conductance: 0.1}}
	}, model.DefaultStartTimeNS)
	w.StateValue("e-1", "a", model.DefaultStartTimeNS+100e9)
	if level := w.entities["e-1"].states["a"].x; level > 10 {
		t.Fatalf("the committed level %v exceeds the state's max 10", level)
	}
}
