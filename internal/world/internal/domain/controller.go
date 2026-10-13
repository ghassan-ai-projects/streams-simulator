package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type controllerState struct {
	integral float64
	previous float64
}

type controllerTerms struct {
	err, proportional, derivative, seconds float64
}

func (w *World) stepController(ent *Entity, dyn *model.Dynamics, s *stateValue, at, dt int64) {
	c := dyn.F1.PID
	pv := w.stateAt(ent.ID, c.PV, at)
	state := s.controllerOf(pv)
	terms := controllerTermsOf(c, w.stateAt(ent.ID, c.SP, at), pv, state.previous, float64(dt)/secondsPerNS)
	state.previous = pv
	if c.Manual != "" && w.stateAt(ent.ID, c.Manual, at) >= 0.5 {
		s.x = clamp(w.stateAt(ent.ID, c.ManualOutput, at), dyn.F1.Clamp)
		state.integral = s.x - terms.proportional - terms.derivative
		return
	}
	s.x = state.integrate(c, dyn.F1.Clamp, terms)
}

func (s *stateValue) controllerOf(pv float64) *controllerState {
	if s.controller == nil {
		s.controller = &controllerState{integral: s.x, previous: pv}
	}
	return s.controller
}

func (c *controllerTerms) output(integral float64) float64 {
	return c.proportional + integral + c.derivative
}

func (state *controllerState) integrate(c *model.Controller, limits *model.Clamp, terms controllerTerms) float64 {
	integral := state.integral + c.Ki*terms.err*terms.seconds
	if !windsUp(terms.output(integral), terms.err, limits) {
		state.integral = integral
	}
	return clamp(terms.output(state.integral), limits)
}

func windsUp(output, err float64, limits *model.Clamp) bool {
	return (limits.Max != nil && output > *limits.Max && err > 0) || (limits.Min != nil && output < *limits.Min && err < 0)
}

func controllerTermsOf(c *model.Controller, sp, pv, previous, seconds float64) controllerTerms {
	err := c.Sign() * (sp - pv)
	return controllerTerms{err: err, proportional: c.Kp * err, derivative: -c.Sign() * c.Kd * (pv - previous) / seconds, seconds: seconds}
}
