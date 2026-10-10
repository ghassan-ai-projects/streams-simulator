package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) stepSpecialForm(ent *Entity, dyn *model.Dynamics, s *stateValue, at, dt int64, seconds float64) bool {
	switch dyn.F1.Form {
	case "saturation", "hysteresis":
		w.stepBoundedForm(ent, dyn, s, at)
	case "dead_time":
		w.stepDeadTime(ent, dyn, s, at, dt, seconds)
	case "rc_network":
		w.stepRCNetwork(ent, dyn, s, at, seconds)
	default:
		return false
	}
	return true
}

func (w *World) stepBoundedForm(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64) {
	input := w.inputValue(ent, dyn, at)
	var value float64
	if dyn.F1.Form == "saturation" {
		value = clamp(dyn.F1.Gain*input, dyn.F1.Clamp)
	} else {
		value = s.hysteresisValue(dyn.F1, input)
	}
	s.x = clamp(value, dyn.F1.Clamp)
}

func (s *stateValue) hysteresisValue(f1 *model.F1Dyn, input float64) float64 {
	value := s.x
	if input >= f1.ThresholdHigh && !s.hysteresis {
		value, s.hysteresis = 1, true
	}
	if input <= f1.ThresholdLow && s.hysteresis {
		value, s.hysteresis = 0, false
	}
	return value
}

func (w *World) stepDeadTime(ent *Entity, dyn *model.Dynamics, s *stateValue, at, dt int64, seconds float64) {
	f1 := dyn.F1
	input := w.inputValue(ent, dyn, at)
	if f1.TimeConstantS > 0 {
		delayed := s.delayedValue(input)
		derivative := (f1.Gain*delayed - s.x) / f1.TimeConstantS
		s.x = clamp(s.x+seconds*derivative, f1.Clamp)
	} else {
		s.x = clamp(f1.Gain*input, f1.Clamp)
	}
	s.recordDelay(input, f1, dt)
}

func (w *World) stepRCNetwork(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64, seconds float64) {
	f1 := dyn.F1
	if f1.TimeConstantS <= 0 || f1.TimeConstant2S <= 0 {
		return
	}
	input := w.inputValue(ent, dyn, at)
	s.integrateRCStages(f1, input, seconds)
}

func (s *stateValue) integrateRCStages(f1 *model.F1Dyn, input, dt float64) {
	// Two cascaded lags: aux is the first stage, x the second.
	tau1, tau2 := f1.TimeConstantS, f1.TimeConstant2S
	a0, x0 := s.aux, s.x
	fa := func(a float64) float64 { return (f1.Gain*input - a) / tau1 }
	fx := func(a, x float64) float64 { return (a - x) / tau2 }
	ka1, kx1 := fa(a0), fx(a0, x0)
	ka2, kx2 := fa(a0+0.5*dt*ka1), fx(a0+0.5*dt*ka1, x0+0.5*dt*kx1)
	ka3, kx3 := fa(a0+0.5*dt*ka2), fx(a0+0.5*dt*ka2, x0+0.5*dt*kx2)
	ka4, kx4 := fa(a0+dt*ka3), fx(a0+dt*ka3, x0+dt*kx3)
	s.aux = clamp(a0+dt/6*(ka1+2*ka2+2*ka3+ka4), f1.Clamp)
	s.x = clamp(x0+dt/6*(kx1+2*kx2+2*kx3+kx4), f1.Clamp)
}

func (w *World) stepFirstOrder(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64, seconds float64) {
	input := w.inputValue(ent, dyn, at)
	k1 := f1Derivative(dyn.F1, s.x, input)
	k2 := f1Derivative(dyn.F1, s.x+0.5*seconds*k1, input)
	k3 := f1Derivative(dyn.F1, s.x+0.5*seconds*k2, input)
	k4 := f1Derivative(dyn.F1, s.x+seconds*k3, input)
	next := s.x + seconds/6*(k1+2*k2+2*k3+k4)
	s.x = clamp(next, dyn.F1.Clamp)
}

func lagDerivative(f1 *model.F1Dyn, value, input float64) float64 {
	if f1.TimeConstantS <= 0 {
		return 0
	}
	return (f1.Gain*input - value) / f1.TimeConstantS
}

func thresholdDerivative(f1 *model.F1Dyn, input float64) float64 {
	if f1.Direction == "below" {
		return f1.Gain * math.Max(0, f1.Threshold-input)
	}
	return f1.Gain * math.Max(0, input-f1.Threshold)
}
