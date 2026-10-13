package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func stochasticWorld(t *testing.T) *World {
	t.Helper()
	w := constWorld(t, func(s *model.DomainSpec) {
		s.Kinds[0].Faults[0].Onset = model.FaultOnset{Shape: "stochastic", RatePerHour: 0.01}
		s.Kinds[0].Faults[0].Affects = []model.FaultEffect{{State: "u", Delta: 3}}
	}, model.DefaultStartTimeNS)
	if _, err := w.InjectFault("e-1", "f1", model.DefaultStartTimeNS+60e9, nil); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestAStochasticFaultReadsTheSameWhateverTheReadOrder(t *testing.T) {
	t.Parallel()
	at := func(minutes int64) int64 { return model.DefaultStartTimeNS + minutes*60e9 }
	silent := stochasticWorld(t)
	early, late := silent.StateValue("e-1", "u", at(30)), silent.StateValue("e-1", "x", at(120))
	busy := stochasticWorld(t)
	for m := int64(0); m <= 120; m++ {
		busy.StateValue("e-1", "u", at(m))
	}
	backwards := stochasticWorld(t)
	backwards.StateValue("e-1", "u", at(240))
	if got := busy.StateValue("e-1", "x", at(120)); got != late {
		t.Fatalf("frequent reads changed x(2 h): %v, silent %v", got, late)
	}
	if got := backwards.StateValue("e-1", "u", at(30)); got != early {
		t.Fatalf("a later read changed u(30 min): %v, alone %v", got, early)
	}
	if early == 2 || late == 0 {
		t.Fatalf("the fault had no effect: u %v, x %v", early, late)
	}
}

func TestReinjectingAStochasticFaultStartsItsOwnWalk(t *testing.T) {
	t.Parallel()
	a, b := stochasticWorld(t), stochasticWorld(t)
	if _, err := b.InjectFault("e-1", "f1", model.DefaultStartTimeNS+7200e9, nil); err != nil {
		t.Fatal(err)
	}
	at := model.DefaultStartTimeNS + 3600e9
	if a.StateValue("e-1", "u", at) != b.StateValue("e-1", "u", at) {
		t.Fatal("a later injection changed the walk of an earlier one")
	}
}
