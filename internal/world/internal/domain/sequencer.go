package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (w *World) stepSequencer(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64) {
	f1, p := dyn.F1, s.protectionOf()
	if !p.seen {
		p.seen, p.until = true, at
	}
	if at-p.until < int64(f1.HoldS*secondsPerNS) {
		return
	}
	switch {
	case s.x < float64(f1.Stages) && w.conditionHolds(ent.ID, f1.StageUp, at):
		s.x, p.until = s.x+1, at
	case s.x > 0 && w.conditionHolds(ent.ID, f1.StageDown, at):
		s.x, p.until = s.x-1, at
	}
}
