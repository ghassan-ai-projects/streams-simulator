package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (w *World) scaledObservation(ent *Entity, ch *model.Channel, reading float64, t int64) float64 {
	return reading * w.observationFactor(ent, ch, t, model.ObservationScale)
}

func (w *World) observationFactor(ent *Entity, ch *model.Channel, t int64, form string) float64 {
	factor := 1.0
	for _, effect := range w.actingObservationFaults(ent, ch, t) {
		if o := effect.fault.fault.Observation; o.Form == form {
			factor *= 1 + (*o.Factor-1)*effect.level
		}
	}
	return factor
}

func (w *World) heldObservation(ent *Entity, ch *model.Channel, t int64) (float64, bool) {
	var held *observationEffect
	for _, effect := range w.actingObservationFaults(ent, ch, t) {
		form := effect.fault.fault.Observation.Form
		if (form == model.ObservationStuck || form == model.ObservationFreeze) && (held == nil || effect.fault.onsetNS >= held.fault.onsetNS) {
			held = &effect
		}
	}
	if held == nil {
		return 0, false
	}
	return w.heldValue(ent, ch, held.fault), true
}

func (w *World) heldValue(ent *Entity, ch *model.Channel, active *activeFault) float64 {
	if o := active.fault.Observation; o.Form == model.ObservationStuck {
		return quantize(*o.Value, ch.Resolution)
	}
	if active.frozen == nil {
		frozen := w.unheldQuietObservation(ent, ch, active.onsetNS)
		active.frozen = &frozen
	}
	return *active.frozen
}

func (w *World) captureFrozen(active *activeFault) {
	if ent := w.entities[active.entity]; ent != nil {
		if ch := ent.kind.ChannelDecl(active.fault.Observation.Channel); ch != nil {
			w.heldValue(ent, ch, active)
		}
	}
}
