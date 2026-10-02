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
	v := w.naturalValue(ent, state, t)
	v, lastKick, anyKick := w.applyStateKicks(v, entity, state, t)
	lastFault := int64(-1)
	anyFault := false
	for _, f := range w.faultsByState[state] {
		if f.entity != entity || !f.activeAt(t) {
			continue
		}
		anyFault = true
		if f.onsetNS > lastFault {
			lastFault = f.onsetNS
		}
	}
	if anyFault && (!anyKick || lastFault > lastKick) {
		// The fault is the most recent driver: it rules the state.
		v = w.naturalValue(ent, state, t)
		for _, f := range w.faultsByState[state] {
			if f.entity != entity || !f.activeAt(t) {
				continue
			}
			add, mult := f.contributionAt(state, t, w.substream(faultSubstream(f, state)), w.dtFor(state))
			if mult != 0 {
				v = v*(1+mult) + add
			} else {
				v += add
			}
		}
	}
	v = w.clampState(ent, state, v)
	return v
}

func (w *World) applyStateKicks(v float64, entity, state string, t int64) (float64, int64, bool) {
	lastKick := int64(-1)
	anyKick := false
	lastAssignment := (*kick)(nil)
	for _, k := range w.kicks[driverKey{entity: entity, state: state}] {
		if t <= k.startNS {
			continue
		}
		anyKick = true
		if k.startNS > lastKick {
			lastKick = k.startNS
		}
		if k.assign && (lastAssignment == nil || k.startNS > lastAssignment.startNS) {
			lastAssignment = k
		}
	}
	if lastAssignment != nil {
		v = lastAssignment.valueAt(t)
		for _, k := range w.kicks[driverKey{entity: entity, state: state}] {
			if k.assign || k.startNS <= lastAssignment.startNS || t <= k.startNS {
				continue
			}
			v += k.valueAt(t)
		}
	} else {
		for _, k := range w.kicks[driverKey{entity: entity, state: state}] {
			if t > k.startNS {
				v += k.valueAt(t)
			}
		}
	}
	return v, lastKick, anyKick
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
	cap := 1
	if f1.DeadTimeS > 0 && dt > 0 {
		cap = int(math.Ceil(f1.DeadTimeS / (float64(dt) / secondsPerNS)))
	}
	if cap < 1 {
		cap = 1
	}
	if len(s.delayed) < cap {
		s.delayed = make([]float64, cap)
	}
	s.delayed[s.delayHead] = u
	s.delayHead = (s.delayHead + 1) % cap
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
	dyn := w.DynamicsFor(state)
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
