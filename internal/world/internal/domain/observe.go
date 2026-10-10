package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Reading computes the noise-free observation of a channel at time t
// (the observability solver's signal source). Public because the truth
// package diffs faulted vs clean readings.
func (w *World) Reading(entityID, channelName string, t int64) float64 {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return 0
	}
	ch := w.Spec.Channel(channelName)
	if ch == nil {
		return 0
	}
	// Reuse the observation path; the solver builds a Noiseless world.
	return w.observe(ent, ch, ent.ensureChannel(channelName), t)
}

// processEmission emits one native event for (entity, channel) at time t.
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
	if w.emitter != nil {
		w.emitter(event)
	}
	w.emittedThisAdvance++
}

// serializeObservedTime turns the link-delay candidate into a strict total
// order in native emission order. Equal candidates receive a one-nanosecond
// deterministic tiebreak, and a candidate that would move backwards is
// advanced past the previous timestamp. The adjustment is deliberately in
// the world layer so every adapter sees the same ordered observed_time.
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
	ch := w.Spec.Channel(channel)
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
	emit, value := w.reading(ent, ch, cs, t)
	if emit {
		delay := w.linkDelay(entityID, ch, t)
		observed := w.serializeObservedTime(t + int64(delay*secondsPerNS))
		w.publishNativeEvent(ent, ch, value, t, observed)
	}
	w.scheduleNextEmission(entityID, channelName, ch, t)
}

// StateValue exposes a hidden state to the director only (solver and truth).
func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.stateAt(entity, state, t)
}
