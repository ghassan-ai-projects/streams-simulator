package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (w *World) advanceEntity(ent *Entity, t int64) {
	horizon := historyHorizon(ent)
	for dyn := nextStep(ent, t); dyn != nil; dyn = nextStep(ent, t) {
		w.commitStep(ent, dyn, horizon)
	}
}

func nextStep(ent *Entity, t int64) *model.Dynamics {
	var next *model.Dynamics
	var from int64
	for i := range ent.kind.Dynamics {
		dyn := &ent.kind.Dynamics[i]
		s := ent.states[dyn.Target]
		if dyn.Tier == "F0" || s == nil || s.lastStep+stepNS(dyn) > t {
			continue
		}
		if next == nil || s.lastStep < from {
			next, from = dyn, s.lastStep
		}
	}
	return next
}

func historyHorizon(ent *Entity) int64 {
	var steps int64
	for i := range ent.kind.Dynamics {
		steps += stepNS(&ent.kind.Dynamics[i])
	}
	return 2 * steps
}

func (w *World) commitStep(ent *Entity, dyn *model.Dynamics, horizon int64) {
	if isStore(dyn) {
		w.commitStores(ent, horizon)
		return
	}
	s := ent.states[dyn.Target]
	s.remember(horizon)
	w.rk4Step(ent, dyn, s, s.lastStep, stepNS(dyn))
}

func (s *stateValue) remember(horizon int64) {
	keepFrom := s.lastStep - horizon
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
	return dyn.F1 != nil && (dyn.F1.Form == model.FormTrip || dyn.F1.Form == model.FormAntiRecycle || dyn.F1.Form == model.FormSequencer)
}

func isStore(dyn *model.Dynamics) bool { return dyn.F1 != nil && dyn.F1.Form == model.FormStore }
