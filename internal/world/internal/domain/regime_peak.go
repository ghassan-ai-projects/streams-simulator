package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type peakState struct {
	sum      float64
	samples  int
	previous float64
}

func (w *World) stepRegime(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64, seconds float64) {
	regime := dyn.F1.Below
	if w.stateAt(ent.ID, dyn.F1.Selector, at) >= dyn.F1.Threshold {
		regime = dyn.F1.Above
	}
	lag := &model.F1Dyn{Form: "first_order_lag", Gain: regime.Gain, TimeConstantS: regime.TimeConstantS, Inputs: dyn.F1.Inputs, Clamp: dyn.F1.Clamp}
	w.stepFirstOrder(ent, &model.Dynamics{Target: dyn.Target, F1: lag}, s, at, seconds)
}

func (w *World) stepPeak(ent *Entity, dyn *model.Dynamics, s *stateValue, at, dt int64) {
	p := s.peakOf(dyn.F1)
	p.sum += w.inputValue(ent, dyn, at)
	p.samples++
	end := at + dt
	if crosses(at, end, dyn.F1.BlockS) {
		s.x = max(s.x, p.sum/float64(p.samples))
		p.sum, p.samples = 0, 0
	}
	if crosses(at, end, dyn.F1.PeriodS) {
		p.previous, s.x = s.x, 0
	}
}

func crosses(from, to int64, everyS float64) bool {
	every := int64(math.Round(everyS * secondsPerNS))
	return to/every != from/every
}

func (s *stateValue) peakOf(f1 *model.F1Dyn) *peakState {
	if s.peak == nil {
		s.peak = &peakState{previous: f1.PreviousPeak}
	}
	return s.peak
}

func billedPeak(f1 *model.F1Dyn, s *stateValue) float64 {
	previous := f1.PreviousPeak
	if s.peak != nil {
		previous = s.peak.previous
	}
	return max(s.x, f1.Ratchet*previous)
}
