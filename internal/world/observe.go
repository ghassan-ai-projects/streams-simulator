package world

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
	cs := ent.channels[channelName]
	if cs == nil {
		cs = &channelRunState{}
		ent.channels[channelName] = cs
	}
	// Reuse the observation path with noise suppressed: the world must be
	// built with Noiseless for a deterministic signal.
	return w.observe(ent, ch, cs, t)
}

// processEmission emits one native event for (entity, channel) at time t.
func (w *World) processEmission(entityID, channelName string, t int64) {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return
	}
	ch := w.Spec.Channel(channelName)
	if ch == nil {
		return
	}
	cs := ent.channels[channelName]
	if cs == nil {
		return
	}

	// Producer availability: during a down period the producer is silent.
	if ch.Availability != nil {
		if err := w.updateAvailability(ent, ch, cs, t); err != nil {
			return
		}
		if cs.availDown {
			// Silent while down; check again when the producer returns.
			w.schedule(kindEmission, entityID, channelName, cs.availUntil, nil)
			return
		}
	}

	// Decide whether this cadence mode emits at this tick.
	emit, value := w.reading(ent, ch, cs, t)
	if !emit {
		next := w.nextEmission(entityID, ch, t)
		if next > 0 {
			w.schedule(kindEmission, entityID, channelName, next, nil)
		}
		return
	}

	// Link delay supplies the observed-time candidate; the world then
	// serializes it into the native emission order.
	delay := w.linkDelay(entityID, ch, t)
	observed := t + int64(delay*secondsPerNS)
	observed = w.serializeObservedTime(observed)

	w.publishNativeEvent(ent, ch, value, t, observed)

	// Schedule the next emission for this channel.
	next := w.nextEmission(entityID, ch, t)
	if next > 0 {
		w.schedule(kindEmission, entityID, channelName, next, nil)
	}
}

func (w *World) publishNativeEvent(ent *Entity, ch *model.Channel, value any, t, observed int64) {
	entityID, channelName := ent.ID, ch.Name
	if !w.EmitDisabled {
		w.seq++
		ev := model.SimEvent{
			Seq:          w.seq - 1,
			WorldID:      w.ID,
			EntityType:   ent.Type,
			EntityID:     entityID,
			Channel:      channelName,
			EventTime:    model.FormatTime(t),
			ObservedTime: model.FormatTime(observed),
			Unit:         ch.Unit,
		}
		if ch.ValueType != "none" {
			ev.Value = value
		}
		if w.emitter != nil {
			w.emitter(ev)
		}
		w.emittedThisAdvance++
	}

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
