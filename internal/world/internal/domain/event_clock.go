package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) schedulePulse(active *activeFault, durationNS int64) {
	active.clearedNS = active.onsetNS + durationNS
	w.checkPulseEdge(active, active.onsetNS)
	w.checkPulseEdge(active, active.clearedNS)
}

func (w *World) checkPulseEdge(active *activeFault, at int64) {
	ent := w.entities[active.entity]
	if ent == nil || active.fault.Onset.Shape != model.ShapePulse {
		return
	}
	for i := range ent.kind.Channels {
		if ch := &ent.kind.Channels[i]; ch.Cadence.Mode == "event_driven" && affectsObserved(active.fault, ch) {
			w.schedule(kindEventCheck, active.entity, ch.Name, max(at, w.clockNS), nil)
		}
	}
}

func affectsObserved(fault *model.Fault, ch *model.Channel) bool {
	return slices.ContainsFunc(fault.Affects, func(effect model.FaultEffect) bool {
		return effect.State == ch.Observes || effect.State == ch.Cadence.TriggerState
	})
}

func (w *World) pulseInstants(entityID string, from, to int64) []int64 {
	var instants []int64
	for _, id := range w.faultOrder {
		if active := w.faultsByID[id]; active.entity == entityID && active.fault.Onset.Shape == model.ShapePulse {
			instants = appendWithin(instants, from, to, active.onsetNS, active.clearedNS)
		}
	}
	slices.Sort(instants)
	return slices.Compact(instants)
}

func (w *World) processEventCheck(entityID, channelName string, t int64) {
	ent, ch, cs := w.emissionChannel(entityID, channelName)
	if cs == nil || w.channelDown(ent, ch, cs, t) {
		return
	}
	if emit, value := w.reading(ent, ch, cs, t); emit {
		delay := w.linkDelay(entityID, ch, t)
		w.publishNativeEvent(ent, ch, value, t, w.serializeObservedTime(t+int64(delay*secondsPerNS)))
	}
}

func (w *World) channelDown(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) bool {
	if ch.Availability == nil {
		return false
	}
	return w.updateAvailability(ent, ch, cs, t) != nil || cs.availDown
}

func appendWithin(instants []int64, from, to int64, times ...int64) []int64 {
	for _, t := range times {
		if t > from && t < to {
			instants = append(instants, t)
		}
	}
	return instants
}
