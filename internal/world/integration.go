package world

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	return w.naturalStateValue(ent, stateName, s, w.DynamicsFor(stateName), t)
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
	value := f0.Baseline + f0.TrendPerHour*hours + f0.DriftPerHour*hours
	return seasonalValue(f0, value, t-w.StartNS)
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
	if dyn.F1 == nil {
		// Unsupported integrated forms retain their current value.
		s.lastStep = t + dt
		return
	}
	dtF := float64(dt) / secondsPerNS
	if !w.stepSpecialForm(ent, dyn, s, t, dt, dtF) {
		w.stepFirstOrder(ent, dyn, s, t, dtF)
	}
	s.lastStep = t + dt
}

func f1Derivative(f1 *model.F1Dyn, x, u float64) float64 {
	switch f1.Form {
	case "first_order_lag":
		return lagDerivative(f1, x, u)
	case "integrator":
		return f1.Gain * u
	case "threshold_integrator":
		return thresholdDerivative(f1, u)
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

func (w *World) naturalStateValue(ent *Entity, name string, s *stateValue, dyn *model.Dynamics, at int64) float64 {
	switch {
	case dyn == nil:
		return w.initial(ent, name)
	case dyn.Tier == "F0":
		return w.f0Value(dyn, at)
	default:
		w.integrateF1(ent, dyn, at)
		return s.x
	}
}

func seasonalValue(f0 *model.F0Dyn, value float64, elapsed int64) float64 {
	for _, season := range f0.Seasonality {
		if season.PeriodS <= 0 {
			continue
		}
		phase := season.PhaseRad
		value += season.Amplitude * math.Sin(2*math.Pi*float64(elapsed)/secondsPerNS/season.PeriodS+phase)
	}
	return value
}
