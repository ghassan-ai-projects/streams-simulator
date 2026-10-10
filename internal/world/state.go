package world

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// stateAt is the full observable hidden value of a state at time t:
// natural dynamics plus the contributions of the state's drivers.
//
// Faults and effect kicks are both drivers of the same state, and the
// most recently activated driver rules: a fault stops a running aerator
// (kick active, fault newer), and a later start_aerator fixes it (kick
// newer than the fault). Without a recency rule the domain's expected
// responses could never work — a permanent additive -1.0 fault would either
// cancel a +1.0 kick forever or be itself uncancellable.
func (w *World) stateAt(entity, state string, t int64) float64 {
	ent := w.entities[entity]
	if ent == nil {
		return 0
	}
	value := w.naturalValue(ent, state, t)
	value, lastKick, anyKick := w.applyStateKicks(value, entity, state, t)
	lastFault, anyFault := w.latestStateFault(entity, state, t)
	if anyFault && (!anyKick || lastFault > lastKick) {
		// The newer fault rules; discard contributions from older kicks.
		value = w.applyStateFaults(w.naturalValue(ent, state, t), entity, state, t)
	}
	return w.clampState(ent, state, value)
}

func (w *World) applyStateKicks(v float64, entity, state string, t int64) (float64, int64, bool) {
	kicks := w.kicks[driverKey{entity: entity, state: state}]
	drivers := latestKickDrivers(kicks, t)
	if drivers.assignment != nil {
		v = applyAssignedKicks(kicks, drivers.assignment, t)
	} else {
		v = applyAdditiveKicks(v, kicks, t)
	}
	return v, drivers.last, drivers.any
}

// activeAt reports whether the fault's contribution is in force at t.
func (f *activeFault) activeAt(t int64) bool {
	if t < f.onsetNS {
		return false
	}
	if f.clearedNS > 0 && t >= f.clearedNS {
		return false
	}
	return true
}

// shadowValue is the state value a confirmation channel reports under a
// silent_no_effect shadow: the real value plus the shadow kicks.
func (w *World) shadowValue(entity, state string, t int64) float64 {
	v := w.stateAt(entity, state, t)
	for _, k := range w.shadow[driverKey{entity: entity, state: state}] {
		v += k.valueAt(t)
	}
	return v
}

func (w *World) clampState(ent *Entity, state string, v float64) float64 {
	for i := range w.Spec.Spec.State {
		st := &w.Spec.Spec.State[i]
		if st.Name != state {
			continue
		}
		if st.Min != 0 || st.Max != 0 {
			return clamp(v, &model.Clamp{Min: fptr(st.Min), Max: fptr(st.Max)})
		}
		return v
	}
	return v
}

func fptr(f float64) *float64 { return &f }

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

// recordDelay pushes u into the dead-time ring buffer.
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

// delayedValue reads the ring buffer entry dead_time behind.
func (s *stateValue) delayedValue(current float64) float64 {
	if !s.delayFull {
		return current
	}
	return s.delayed[s.delayHead]
}

// dtFor returns the integration step for a state (ns), or 0.
func (w *World) dtFor(state string) int64 {
	dyn := w.dynamicsFor(state)
	if dyn == nil || (dyn.Tier != "F1" && dyn.Tier != "F2") {
		return 0
	}
	if dyn.DTMs <= 0 {
		return secondsPerNS
	}
	return dyn.DTMs * 1e6
}

func faultSubstream(f *activeFault, state string) string {
	return "fault/" + f.fault.ID + "/" + state
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
