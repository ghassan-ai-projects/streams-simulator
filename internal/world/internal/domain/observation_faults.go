package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var qualityRank = map[string]int{"": 0, "suspect": 1, "stale": 2, "out_of_range": 3, "disconnected": 4}

func (w *World) registerObservationFault(active *activeFault) {
	o := active.fault.Observation
	if o == nil {
		return
	}
	key := driverKey{entity: active.entity, state: o.Channel}
	w.faultsByChannel[key] = append(w.faultsByChannel[key], active)
	if o.Form == model.ObservationFreeze {
		w.schedule(kindObservationCapture, active.entity, o.Channel, active.onsetNS, active)
	}
}

func (w *World) actingObservationFaults(ent *Entity, ch *model.Channel, t int64) []observationEffect {
	var acting []observationEffect
	for _, active := range w.faultsByChannel[driverKey{entity: ent.ID, state: ch.Name}] {
		key := "observation/" + ch.Name
		if level := active.envelopeAt(key, t, w.substream(faultSubstream(active, key)), secondsPerNS) * active.severity; level > 0 {
			acting = append(acting, observationEffect{fault: active, level: level})
		}
	}
	return acting
}

type observationEffect struct {
	fault *activeFault
	level float64
}

func (w *World) droppedOut(ent *Entity, ch *model.Channel, t int64) bool {
	for _, effect := range w.actingObservationFaults(ent, ch, t) {
		if effect.fault.fault.Observation.Form == model.ObservationDropout {
			return true
		}
	}
	return false
}

func (w *World) reportedQuality(ent *Entity, ch *model.Channel, t int64) string {
	worst := ""
	for _, effect := range w.actingObservationFaults(ent, ch, t) {
		if reported := effect.fault.fault.Observation.Reported(); qualityRank[reported] > qualityRank[worst] {
			worst = reported
		}
	}
	return worst
}

func (w *World) ReadingNoise(entityID, channelName string, t int64) float64 {
	ent, ch := w.liveChannel(entityID, channelName)
	if ch == nil {
		return 0
	}
	return ch.Noise.Sigma * w.observationFactor(ent, ch, t, model.ObservationNoise)
}

func (w *World) ReadingDropped(entityID, channelName string, t int64) bool {
	ent, ch := w.liveChannel(entityID, channelName)
	return ch != nil && w.droppedOut(ent, ch, t)
}

func (w *World) liveChannel(entityID, channelName string) (*Entity, *model.Channel) {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return nil, nil
	}
	return ent, ent.kind.ChannelDecl(channelName)
}
