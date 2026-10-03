package world

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// reading computes the observed (pre-quantization) value and whether this
// tick emits, per the channel's cadence mode.
func (w *World) reading(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) (bool, any) {
	switch ch.ValueType {
	case "none":
		return w.cadenceEmits(ch, cs, t, 0), nil
	case "string":
		return w.stringReading(ent, ch, cs, t)
	default:
		return w.numericReading(ent, ch, cs, t)
	}
}

// observe applies the channel's observation function to the hidden state.
func (w *World) observe(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) float64 {
	value := w.stateAt(ent.ID, ch.Observes, t)
	if w.isConfirmationChannel(ch.Name) {
		value = w.shadowValue(ent.ID, ch.Observes, t)
	}
	reading := w.biasedObservation(ent, ch, value, t)
	reading = w.noisyObservation(ent, ch, reading, t)
	if drift := ch.Drift; drift != nil && drift.RatePerHour != 0 {
		reading += w.drift(ent, ch, cs, t, drift)
	}
	return quantize(reading, ch.Resolution)
}

// drift advances the reading drift (linear or random walk).
func (w *World) drift(ent *Entity, ch *model.Channel, cs *channelRunState, t int64, d *model.Drift) float64 {
	if cs.walkLastNS == 0 {
		cs.walkLastNS = t
		return d.RatePerHour * (float64(t-w.StartNS) / secondsPerNS / 3600)
	}
	hours := float64(t-cs.walkLastNS) / secondsPerNS / 3600
	cs.walkLastNS = t
	if d.Model == "random_walk" {
		return w.randomWalkDrift(ent, ch, cs, d, hours)
	}
	return d.RatePerHour * (float64(t-w.StartNS) / secondsPerNS / 3600)
}

func (w *World) randomWalkDrift(ent *Entity, ch *model.Channel, cs *channelRunState, drift *model.Drift, hours float64) float64 {
	rng := w.substream(ent.ID + "/" + ch.Name + "/drift")
	step := drift.RatePerHour*hours + math.Abs(drift.RatePerHour)*math.Sqrt(hours)*rng.Norm()
	cs.walk += step
	return cs.walk
}
