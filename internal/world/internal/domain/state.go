package domain

import (
	"math"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) stateAt(entity, state string, t int64) float64 {
	ent := w.entities[entity]
	if ent == nil {
		return 0
	}
	value := w.naturalValue(ent, state, t)
	value, lastKick, anyKick := w.applyStateKicks(value, entity, state, t)
	lastFault, anyFault := w.latestStateFault(entity, state, t)
	if anyFault && (!anyKick || lastFault > lastKick) {

		value = w.applyStateFaults(w.naturalValue(ent, state, t), entity, state, t)
	}
	return w.clampState(ent, state, value)
}

func (w *World) applyStateKicks(v float64, entity, state string, t int64) (float64, int64, bool) {
	kicks := w.kicks[driverKey{entity: entity, state: state}]
	drivers := latestKickDrivers(kicks, t)
	return drivenValue(v, kicks, t, t), drivers.last, drivers.any
}

func (f *activeFault) activeAt(t int64) bool {
	if t < f.onsetNS {
		return false
	}
	if f.clearedNS > 0 && t >= f.clearedNS {
		return false
	}
	return true
}

func (w *World) shadowValue(entity, state string, t int64) float64 {
	shadow := w.shadow[driverKey{entity: entity, state: state}]
	return drivenValue(w.stateAt(entity, state, t), shadow, t, t)
}

func (w *World) clampState(ent *Entity, state string, v float64) float64 {
	if st := ent.kind.StateDecl(state); st != nil {
		return clampToBounds(v, st.Min, st.Max)
	}
	return v
}

func clampToBounds(v float64, lower, upper *float64) float64 {
	if lower != nil {
		v = max(v, *lower)
	}
	if upper != nil {
		v = min(v, *upper)
	}
	return v
}

func clamp(v float64, c *model.Clamp) float64 {
	if c == nil {
		return v
	}
	if c.Min != nil && v < *c.Min {
		return *c.Min
	}
	if c.Max != nil && v > *c.Max {
		return *c.Max
	}
	return v
}

func (s *stateValue) recordDelay(u float64, f1 *model.F1Dyn, dt int64) {
	capacity := delayCapacity(f1, dt)
	if len(s.delayed) < capacity {
		s.delayed = make([]float64, capacity)
	}
	s.delayed[s.delayHead] = u
	s.delayHead = (s.delayHead + 1) % capacity
	if s.delayHead == 0 {
		s.delayFull = true
	}
}

func (s *stateValue) delayedValue(current float64) float64 {
	if !s.delayFull {
		return current
	}
	return s.delayed[s.delayHead]
}

func dtFor(kind *model.Kind, state string) int64 {
	dyn := kind.DynamicsOf(state)
	if dyn == nil || (dyn.Tier != "F1" && dyn.Tier != "F2") {
		return 0
	}
	if dyn.DTMs <= 0 {
		return secondsPerNS
	}
	return dyn.DTMs * 1e6
}

func faultSubstream(f *activeFault, state string) string {
	return "fault/" + f.fault.ID + "/" + f.entity + "/" + strconv.FormatInt(f.onsetNS, 10) + "/" + state
}

func delayCapacity(f1 *model.F1Dyn, dt int64) int {
	capacity := 1
	if f1.DeadTimeS > 0 && dt > 0 {
		capacity = int(math.Ceil(f1.DeadTimeS / (float64(dt) / secondsPerNS)))
	}
	if capacity < 1 {
		capacity = 1
	}
	return capacity
}
