package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) Reading(entityID, channelName string, t int64) float64 {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return 0
	}
	ch := ent.kind.ChannelDecl(channelName)
	if ch == nil {
		return 0
	}
	return w.noiseFreeObservation(ent, ch, t)
}

func (w *World) noiseFreeObservation(ent *Entity, ch *model.Channel, t int64) float64 {
	if w.noiseless {
		return w.observe(ent, ch, ent.ensureChannel(ch.Name), t)
	}
	return w.quietObservation(ent, ch, t)
}

func (w *World) quietObservation(ent *Entity, ch *model.Channel, t int64) float64 {
	if held, ok := w.heldObservation(ent, ch, t); ok {
		return held
	}
	return w.unheldQuietObservation(ent, ch, t)
}

func (w *World) unheldQuietObservation(ent *Entity, ch *model.Channel, t int64) float64 {
	value := w.stateAt(ent.ID, ch.Observes, t)
	if w.isConfirmationChannel(ent, ch.Name) {
		value = w.shadowValue(ent.ID, ch.Observes, t)
	}
	reading := w.scaledObservation(ent, ch, w.biasedObservation(ent, ch, value, t), t)
	if drift := ch.Drift; drift != nil && drift.RatePerHour != 0 && drift.Model != "random_walk" {
		reading += drift.RatePerHour * (float64(t-w.StartNS) / secondsPerNS / 3600)
	}
	return quantize(reading, ch.Resolution)
}

func (w *World) processEmission(entityID, channelName string, t int64) {
	ent, ch, cs := w.emissionChannel(entityID, channelName)
	if cs == nil {
		return
	}
	if !w.producerAvailable(entityID, channelName, ent, ch, cs, t) {
		return
	}
	w.emitChannelReading(entityID, channelName, ent, ch, cs, t)
}

func (w *World) publishNativeEvent(ent *Entity, ch *model.Channel, value any, t, observed int64) {
	if w.emitDisabled {
		return
	}
	w.seq++
	event := w.nativeEvent(ent, ch, value, t, observed)
	event.Quality = w.reportedQuality(ent, ch, t)
	if w.emitter != nil {
		w.emitter(event)
	}
	w.emittedThisAdvance++
}

func (w *World) serializeObservedTime(candidate int64) int64 {
	if w.hasObservedTimeNS && candidate <= w.lastObservedNS {
		candidate = w.lastObservedNS + 1
	}
	w.lastObservedNS = candidate
	w.hasObservedTimeNS = true
	return candidate
}

func (ent *Entity) ensureChannel(name string) *channelRunState {
	state := ent.channels[name]
	if state == nil {
		state = &channelRunState{}
		ent.channels[name] = state
	}
	return state
}

func (w *World) emissionChannel(entity, channel string) (*Entity, *model.Channel, *channelRunState) {
	ent := w.entities[entity]
	if ent == nil || !ent.alive {
		return nil, nil, nil
	}
	ch := ent.kind.ChannelDecl(channel)
	if ch == nil {
		return nil, nil, nil
	}
	return ent, ch, ent.channels[channel]
}

func (w *World) producerAvailable(entityID, channelName string, ent *Entity, ch *model.Channel, cs *channelRunState, at int64) bool {
	if ch.Availability == nil {
		return true
	}
	if err := w.updateAvailability(ent, ch, cs, at); err != nil {
		return false
	}
	if cs.availDown {
		w.schedule(kindEmission, entityID, channelName, cs.availUntil, nil)
		return false
	}
	return true
}

func (w *World) scheduleNextEmission(entity, channel string, ch *model.Channel, at int64) {
	if w.unobserved {
		return
	}
	next := w.nextEmission(entity, ch, at)
	if next > 0 {
		w.schedule(kindEmission, entity, channel, next, nil)
	}
}

func (w *World) nativeEvent(ent *Entity, ch *model.Channel, value any, at, observed int64) model.SimEvent {
	event := model.SimEvent{Seq: w.seq - 1, WorldID: w.ID, EntityType: ent.Type, EntityID: ent.ID,
		Channel: ch.Name, EventTime: model.FormatTime(at), ObservedTime: model.FormatTime(observed), Unit: ch.Unit}
	if ch.ValueType != "none" {
		event.Value = value
	}
	return event
}

func (w *World) emitChannelReading(entityID, channelName string, ent *Entity, ch *model.Channel, cs *channelRunState, t int64) {
	if w.droppedOut(ent, ch, t) {
		w.scheduleNextEmission(entityID, channelName, ch, t)
		return
	}
	emit, value := w.reading(ent, ch, cs, t)
	if emit {
		delay := w.linkDelay(entityID, ch, t)
		observed := w.serializeObservedTime(t + int64(delay*secondsPerNS))
		w.publishNativeEvent(ent, ch, value, t, observed)
	}
	w.scheduleNextEmission(entityID, channelName, ch, t)
}

func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.stateAt(entity, state, t)
}
