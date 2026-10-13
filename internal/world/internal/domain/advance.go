package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var discreteForms = []string{model.FormTrip, model.FormAntiRecycle, model.FormSequencer, model.FormPID}

func (w *World) advanceEntity(ent *Entity, t int64) {
	w.advanceReaders(ent, t)
	dyn := nextStep(ent, t)
	if dyn == nil {
		return
	}
	keep := retention{horizon: 2 * kindSteps(ent.kind), floor: w.readersFloor(ent)}
	for ; dyn != nil; dyn = nextStep(ent, t) {
		w.commitStep(ent, dyn, keep)
	}
}

func nextStep(ent *Entity, t int64) *model.Dynamics {
	var next *model.Dynamics
	var from int64
	for i := range ent.kind.Dynamics {
		dyn := &ent.kind.Dynamics[i]
		s := ent.states[dyn.Target]
		if !dyn.Integrated() || s == nil || s.lastStep+stepNS(dyn) > t {
			continue
		}
		if next == nil || s.lastStep < from {
			next, from = dyn, s.lastStep
		}
	}
	return next
}

func (w *World) commitStep(ent *Entity, dyn *model.Dynamics, keep retention) {
	if isStore(dyn) {
		w.commitStores(ent, keep)
		return
	}
	s := ent.states[dyn.Target]
	s.remember(keep)
	w.rk4Step(ent, dyn, s, s.lastStep, stepNS(dyn))
}

func (s *stateValue) remember(keep retention) {
	keepFrom := min(s.lastStep-keep.horizon, keep.floor)
	for len(s.history) > 1 && s.history[1].lastStep <= keepFrom {
		s.history = s.history[1:]
	}
	s.history = append(s.history, s.clone())
}

func (s *stateValue) at(t int64) *stateValue {
	for i := len(s.history) - 1; i >= 0 && s.lastStep > t; i-- {
		if s.history[i].lastStep <= t {
			return s.history[i]
		}
	}
	return s
}

func (w *World) valueAt(ent *Entity, dyn *model.Dynamics, t int64) float64 {
	s := ent.states[dyn.Target].at(t)
	if dyn.F1 != nil && dyn.F1.Form == model.FormPeak {
		return billedPeak(dyn.F1, s)
	}
	if s.lastStep >= t || isDiscrete(dyn) {
		return s.x
	}
	if isStore(dyn) {
		return w.storeProbe(ent, dyn.Target, s.lastStep, t)
	}
	probe := s.clone()
	w.rk4Step(ent, dyn, probe, probe.lastStep, t-probe.lastStep)
	return probe.x
}

func isDiscrete(dyn *model.Dynamics) bool {
	return dyn.F1 != nil && slices.Contains(discreteForms, dyn.F1.Form)
}

func isStore(dyn *model.Dynamics) bool { return dyn.F1 != nil && dyn.F1.Form == model.FormStore }
