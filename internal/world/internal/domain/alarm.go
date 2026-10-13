package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type alarmState struct {
	active  bool
	acked   bool
	pending int64
	lastAck float64
}

func (w *World) stepAlarm(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64) {
	a := dyn.F1.Alarm
	ack := w.ackCount(ent, a, at)
	state := s.alarmOf(ack)
	if ack != state.lastAck {
		state.acked, state.lastAck = true, ack
	}
	state.track(a, w.alarmCondition(ent, a, state.active, at), at)
	s.x = state.code(a.Shelve != "" && w.stateAt(ent.ID, a.Shelve, at) >= model.ShelveThreshold)
}

func (s *stateValue) alarmOf(ack float64) *alarmState {
	if s.alarm == nil {
		s.alarm = &alarmState{acked: true, pending: -1, lastAck: ack}
	}
	return s.alarm
}

func (w *World) ackCount(ent *Entity, a *model.Alarm, at int64) float64 {
	if a.Acknowledge == "" {
		return 0
	}
	return w.stateAt(ent.ID, a.Acknowledge, at)
}

func (w *World) alarmCondition(ent *Entity, a *model.Alarm, active bool, at int64) bool {
	limit := w.alarmLimit(ent, a, at)
	margin := 0.0
	if active {
		margin = a.Deadband
	}
	if a.Direction == model.AlarmHigh {
		return w.stateAt(ent.ID, a.Watch, at) > limit-margin
	}
	return w.stateAt(ent.ID, a.Watch, at) < limit+margin
}

func (w *World) alarmLimit(ent *Entity, a *model.Alarm, at int64) float64 {
	if a.LimitState != "" {
		return w.stateAt(ent.ID, a.LimitState, at)
	}
	return *a.Limit
}

func (state *alarmState) track(a *model.Alarm, condition bool, at int64) {
	if condition == state.active {
		state.pending = -1
		return
	}
	if state.pending < 0 {
		state.pending = at
	}
	if float64(at-state.pending)/secondsPerNS >= transitionDelay(a, condition) {
		state.active, state.pending = condition, -1
		state.acked = state.acked && !condition
	}
}

func transitionDelay(a *model.Alarm, activating bool) float64 {
	if activating {
		return a.OnDelayS
	}
	return a.OffDelayS
}

func (state *alarmState) code(shelved bool) float64 {
	switch {
	case shelved:
		return model.AlarmShelved
	case state.active && !state.acked:
		return model.AlarmUnacked
	case state.active:
		return model.AlarmAcked
	case !state.acked:
		return model.AlarmReturned
	}
	return model.AlarmNormal
}
