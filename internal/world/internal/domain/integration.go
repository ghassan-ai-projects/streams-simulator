package domain

import (
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type stateValue struct {
	x          float64
	lastStep   int64
	aux        float64
	delayed    []float64
	delayHead  int
	delayFull  bool
	hysteresis bool
	history    []*stateValue
	peak       *peakState
	protection *protectionState
}

func (w *World) naturalValue(ent *Entity, stateName string, t int64) float64 {
	s := ent.states[stateName]
	if s == nil {
		return 0
	}
	return w.naturalStateValue(ent, stateName, s, ent.kind.DynamicsOf(stateName), t)
}

func (w *World) initial(ent *Entity, stateName string) float64 {
	if st := ent.kind.StateDecl(stateName); st != nil {
		return st.Initial
	}
	return 0
}

func (w *World) f0Value(dyn *model.Dynamics, t int64) float64 {
	f0 := dyn.F0
	if f0 == nil {
		return 0
	}
	hours := float64(t-w.StartNS) / secondsPerNS / 3600
	value := f0.Baseline + f0.TrendPerHour*hours + f0.DriftPerHour*hours
	return seasonalValue(f0, value, t-w.StartNS) + scheduleValue(f0.Schedule, t, w.StartNS)
}

func (w *World) integrateF1(ent *Entity, dyn *model.Dynamics, t int64) float64 {
	w.advanceEntity(ent, t)
	return w.valueAt(ent, dyn, t)
}

func (s *stateValue) clone() *stateValue {
	c := *s
	c.delayed = slices.Clone(s.delayed)
	c.history = nil
	if s.peak != nil {
		peak := *s.peak
		c.peak = &peak
	}
	if s.protection != nil {
		protection := *s.protection
		c.protection = &protection
	}
	return &c
}

func (w *World) rk4Step(ent *Entity, dyn *model.Dynamics, s *stateValue, t, dt int64) {
	if dyn.F1 == nil {

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

func (w *World) inputValue(ent *Entity, dyn *model.Dynamics, t int64) float64 {
	f1 := dyn.F1
	if f1 == nil || len(f1.Inputs) == 0 {
		return 0
	}
	var u float64
	for _, in := range f1.Inputs {
		u += in.Coef * w.inputState(ent, in, t)
	}
	return u
}

func (w *World) naturalStateValue(ent *Entity, name string, s *stateValue, dyn *model.Dynamics, at int64) float64 {
	if value, linked := w.followValue(ent, dyn, at); linked {
		return value
	}
	switch {
	case dyn == nil:
		return w.initial(ent, name)
	case dyn.Tier == "F0":
		return w.f0Value(dyn, at)
	case dyn.Algebraic():
		return w.algebraicValue(ent, dyn, at)
	default:
		return w.integrateF1(ent, dyn, at)
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
