package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

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
	case "rc_network":
		// Two cascaded first-order lags. aux is the first stage and x is
		// the second; both declared time constants affect the output.
		tau1, tau2 := f1.TimeConstantS, f1.TimeConstant2S
		if tau1 <= 0 || tau2 <= 0 {
			s.lastStep = t + dt
			return
		}
		u := w.inputValue(ent, dyn, t)
		a0, x0 := s.aux, s.x
		fa := func(a float64) float64 { return (f1.Gain*u - a) / tau1 }
		fx := func(a, x float64) float64 { return (a - x) / tau2 }
		ka1, kx1 := fa(a0), fx(a0, x0)
		ka2, kx2 := fa(a0+0.5*dtF*ka1), fx(a0+0.5*dtF*ka1, x0+0.5*dtF*kx1)
		ka3, kx3 := fa(a0+0.5*dtF*ka2), fx(a0+0.5*dtF*ka2, x0+0.5*dtF*kx2)
		ka4, kx4 := fa(a0+dtF*ka3), fx(a0+dtF*ka3, x0+dtF*kx3)
		s.aux = clamp(a0+dtF/6*(ka1+2*ka2+2*ka3+ka4), dyn.F1.Clamp)
		s.x = clamp(x0+dtF/6*(kx1+2*kx2+2*kx3+kx4), dyn.F1.Clamp)
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
