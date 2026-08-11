package world

// Hidden-state dynamics: F0 statistical generators, F1 forms integrated with
// fixed-step RK4, effect kicks with physical time constants, and fault
// contributions with declared onset shapes. All of it is a pure function of
// (domain spec, seed, command log) — no wall clock, no shared randomness.

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

// secondsPerNS converts nanoseconds to seconds.
const secondsPerNS = 1e9

// kick is a permanent state contribution that approaches its full delta
// exponentially with a time constant, starting at startNS (after dead time).
type kick struct {
	entity  string
	state   string
	delta   float64
	startNS int64
	tauNS   float64 // 0 = instantaneous
	applied bool    // set when the effect-start event pops
}

// driverKey scopes a physical or shadow driver to one entity and state. A
// state name alone is not an identity: two producers may expose the same
// state while remaining physically independent.
type driverKey struct {
	entity string
	state  string
}

func (k *kick) valueAt(t int64) float64 {
	if t <= k.startNS {
		return 0
	}
	if k.tauNS <= 0 {
		return k.delta
	}
	return k.delta * (1 - math.Exp(-float64(t-k.startNS)/k.tauNS))
}

// activeFault is one injected fault and its current contribution to each
// affected state.
type activeFault struct {
	fault     *model.Fault
	entity    string
	onsetNS   int64
	clearedNS int64 // 0 while active; > 0 once cleared (contribution reverts to 0)
	severity  float64
	// stochastic envelope state, advanced lazily on read.
	walk     float64
	walkStep int64
}

// envelopeAt returns the normalized onset envelope e(t) in [0, 1] for the
// elapsed time since onset. Deterministic: intermittent and stochastic use
// the fault's own substream, so no other injection surface is affected.
func (f *activeFault) envelopeAt(t int64, rng *randutil.SplitMix64, dtNS int64) float64 {
	elapsed := t - f.onsetNS
	if elapsed < 0 {
		return 0
	}
	if f.clearedNS > 0 && t >= f.clearedNS {
		return 0
	}
	shape := f.fault.Onset.Shape
	rate := f.fault.Onset.RatePerHour
	switch shape {
	case "step":
		return 1
	case "ramp":
		e := rate * (float64(elapsed) / secondsPerNS / 3600)
		return math.Min(e, 1)
	case "exponential":
		return 1 - math.Exp(-rate*float64(elapsed)/secondsPerNS/3600)
	case "intermittent":
		period := 3600.0
		if rate > 0 {
			period = 3600 / rate
		}
		periodNS := int64(period * secondsPerNS)
		cycle := elapsed % periodNS
		duty := f.fault.Onset.DutyCycle
		if duty <= 0 {
			duty = 0.5
		}
		if float64(cycle) < duty*float64(periodNS) {
			return 1
		}
		return 0
	case "stochastic":
		if dtNS <= 0 {
			dtNS = 60 * secondsPerNS
		}
		for f.walkStep < t {
			next := f.walkStep + dtNS
			if next > t {
				next = t
			}
			dh := float64(next-f.walkStep) / secondsPerNS / 3600
			f.walk += rate*dh + 0.5*math.Sqrt(rate*dh+1e-12)*math.Abs(rng.Norm())
			f.walkStep = next
		}
		return math.Min(math.Max(f.walk, 0), 1)
	}
	return 0
}

// contributionAt is this fault's additive or multiplicative contribution to
// one affected state at time t, scaled by the injected severity.
func (f *activeFault) contributionAt(state string, t int64, rng *randutil.SplitMix64, dtNS int64) (add, mult float64) {
	for _, a := range f.fault.Affects {
		if a.State != state {
			continue
		}
		e := f.envelopeAt(t, rng, dtNS) * f.severity
		if a.Multiplicative {
			mult += a.Delta * e
		} else {
			add += a.Delta * e
		}
	}
	return add, mult
}

// stateValue holds the per-entity integration state of one hidden state.
type stateValue struct {
	x          float64   // integrated value (F1) or natural value (F0/piecewise)
	lastStep   int64     // last integration time
	aux        float64   // second lag value for rc_network
	delayed    []float64 // dead-time ring buffer
	delayHead  int
	delayFull  bool
	hysteresis bool
}

// naturalValue computes the fault- and kick-free value of a state at time t,
// integrating F1 states forward as needed. Returns the value and the state
// record (so callers can keep the updated integration state).
func (w *World) naturalValue(ent *Entity, stateName string, t int64) float64 {
	s := ent.States[stateName]
	if s == nil {
		return 0
	}
	dyn := w.DynamicsFor(stateName)
	switch {
	case dyn == nil:
		// Piecewise constant: natural value is the initial value.
		return w.initial(ent, stateName)
	case dyn.Tier == "F0":
		return w.f0Value(dyn, t)
	default: // F1
		w.integrateF1(ent, dyn, t)
		return s.x
	}
}

func (w *World) initial(ent *Entity, stateName string) float64 {
	for i := range w.Spec.Spec.State {
		st := &w.Spec.Spec.State[i]
		if st.Name == stateName {
			return st.Initial
		}
	}
	return 0
}

// f0Value evaluates an F0 generator analytically at time t.
func (w *World) f0Value(dyn *model.Dynamics, t int64) float64 {
	f0 := dyn.F0
	if f0 == nil {
		return 0
	}
	hours := float64(t-w.StartNS) / secondsPerNS / 3600
	v := f0.Baseline + f0.TrendPerHour*hours + f0.DriftPerHour*hours
	for _, s := range f0.Seasonality {
		if s.PeriodS <= 0 {
			continue
		}
		phase := s.PhaseRad
		v += s.Amplitude * math.Sin(2*math.Pi*float64(t-w.StartNS)/secondsPerNS/s.PeriodS+phase)
	}
	return v
}

// integrateF1 advances an F1 state from its last step to t in dt-sized RK4
// steps (with one final partial step), reading inputs at step boundaries.
func (w *World) integrateF1(ent *Entity, dyn *model.Dynamics, t int64) {
	s := ent.States[dyn.Target]
	dt := dyn.DTMs * 1e6 // ms -> ns
	if dt <= 0 {
		dt = secondsPerNS
	}
	for s.lastStep < t {
		step := dt
		if s.lastStep+step > t {
			step = t - s.lastStep
		}
		w.rk4Step(ent, dyn, s, s.lastStep, step)
	}
}

func (w *World) rk4Step(ent *Entity, dyn *model.Dynamics, s *stateValue, t, dt int64) {
	f1 := dyn.F1
	if f1 == nil {
		// F2/F3 have no integrated form here (F3 is not_implemented); leave
		// the value where it is.
		s.lastStep = t + dt
		return
	}
	dtF := float64(dt) / secondsPerNS
	switch f1.Form {
	case "saturation", "hysteresis":
		u := w.inputValue(ent, dyn, t)
		var x float64
		if f1.Form == "saturation" {
			x = clamp(f1.Gain*u, dyn.F1.Clamp)
		} else {
			x = s.x
			if u >= f1.ThresholdHigh && !s.hysteresis {
				x, s.hysteresis = 1, true
			}
			if u <= f1.ThresholdLow && s.hysteresis {
				x, s.hysteresis = 0, false
			}
		}
		s.x = clamp(x, dyn.F1.Clamp)
		s.lastStep = t + dt
		return
	case "dead_time":
		u := w.inputValue(ent, dyn, t)
		if f1.TimeConstantS > 0 {
			// first-order lag driven by the delayed input
			delayed := s.delayedValue(u)
			dx := (f1.Gain*delayed - s.x) / f1.TimeConstantS
			s.x = clamp(s.x+dtF*dx, dyn.F1.Clamp)
		} else {
			s.x = clamp(f1.Gain*u, dyn.F1.Clamp)
		}
		s.recordDelay(u, f1, dt)
		s.lastStep = t + dt
		return
	}

	// Standard RK4 for the remaining first-order forms.
	inputs := w.inputValue(ent, dyn, t)
	k1 := f1Derivative(f1, s.x, inputs)
	k2 := f1Derivative(f1, s.x+0.5*dtF*k1, inputs)
	k3 := f1Derivative(f1, s.x+0.5*dtF*k2, inputs)
	k4 := f1Derivative(f1, s.x+dtF*k3, inputs)
	next := s.x + dtF/6*(k1+2*k2+2*k3+k4)
	s.x = clamp(next, dyn.F1.Clamp)
	if f1.Form == "rc_network" {
		// second lag: dm/dt = (g*u - m)/tau1 was k1..k4 of the first;
		// the second lag integrates aux toward the first stage output.
		dmdt := (f1.Gain*inputs - s.aux) / f1.TimeConstantS
		s.aux = clamp(s.aux+dtF*dmdt, dyn.F1.Clamp)
		// x is the second stage: drive it toward the first stage (s.aux).
		s.x = clamp(s.aux, dyn.F1.Clamp)
	}
	s.lastStep = t + dt
}

func f1Derivative(f1 *model.F1Dyn, x, u float64) float64 {
	switch f1.Form {
	case "first_order_lag":
		if f1.TimeConstantS <= 0 {
			return 0
		}
		return (f1.Gain*u - x) / f1.TimeConstantS
	case "integrator":
		return f1.Gain * u
	case "threshold_integrator":
		if f1.Direction == "below" {
			return f1.Gain * math.Max(0, f1.Threshold-u)
		}
		return f1.Gain * math.Max(0, u-f1.Threshold)
	case "rc_network":
		if f1.TimeConstantS <= 0 {
			return 0
		}
		return (f1.Gain*u - x) / f1.TimeConstantS
	}
	return 0
}

// inputValue is the weighted sum of the dynamics' declared inputs at time t,
// evaluated at step boundaries (explicit coupling between states).
func (w *World) inputValue(ent *Entity, dyn *model.Dynamics, t int64) float64 {
	f1 := dyn.F1
	if f1 == nil || len(f1.Inputs) == 0 {
		return 0
	}
	var u float64
	for _, in := range f1.Inputs {
		u += in.Coef * w.stateAt(ent.ID, in.State, t)
	}
	return u
}

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
	lastKick := int64(-1)
	anyKick := false
	for _, k := range w.kicks[driverKey{entity: entity, state: state}] {
		kv := k.valueAt(t)
		if kv == 0 {
			continue
		}
		anyKick = true
		if k.startNS > lastKick {
			lastKick = k.startNS
		}
		v += kv
	}
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
