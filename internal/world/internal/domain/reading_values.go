package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

func (w *World) stringReading(ent *Entity, ch *model.Channel, cs *channelRunState, at int64) (bool, any) {
	if ch.Cadence.Mode != "event_driven" {
		return w.cadenceEmits(ch, cs, at, 0), nil
	}
	current := w.stateAt(ent.ID, ch.Cadence.TriggerState, at)
	changed := !cs.hasSent || math.Abs(current-cs.lastTrigger) >= ch.Resolution
	cs.lastTrigger = current
	if !changed {
		return false, nil
	}
	cs.hasSent = true
	return w.enumReading(ent, ch)
}

func (w *World) numericReading(ent *Entity, ch *model.Channel, cs *channelRunState, at int64) (bool, any) {
	raw := w.observe(ent, ch, cs, at)
	if !w.cadenceEmits(ch, cs, at, raw) {
		return false, nil
	}
	cs.lastSent = raw
	cs.hasSent = true
	return true, typedReading(ch, raw)
}

func typedReading(ch *model.Channel, raw float64) any {
	switch ch.ValueType {
	case "boolean":
		return raw >= 0.5
	case "counter":
		return int64(math.Round(raw))
	default:
		return raw
	}
}

func (w *World) biasedObservation(ent *Entity, ch *model.Channel, value float64, at int64) float64 {
	gain := w.Spec.ChannelGain(ch.Name)
	reading := gain*value + ch.ObservationOffset
	if bias := ch.ObservationBias; bias != nil {
		reading += bias.Coef * w.stateAt(ent.ID, bias.State, at)
	}
	return reading
}

func (w *World) observationSigma(ent *Entity, ch *model.Channel, at int64) float64 {
	sigma := ch.Noise.Sigma
	if bias := ch.ObservationBias; bias != nil && sigma > 0 && !w.noiseless {
		value := math.Min(math.Max(w.stateAt(ent.ID, bias.State, at), 0), 1)
		scale := 1 + (bias.NoiseScaleAtFull-1)*value
		sigma *= scale
	}
	return sigma
}

func (w *World) noisyObservation(ent *Entity, ch *model.Channel, reading float64, at int64) float64 {
	sigma := w.observationSigma(ent, ch, at)
	if sigma > 0 && ch.Noise.Model == "quantization" && !w.noiseless {
		return math.Round(reading/sigma) * sigma
	}
	if sigma > 0 && !w.noiseless {
		rng := w.substream(ent.ID + "/" + ch.Name + "/noise")
		reading += sigma * rng.Norm()
	}
	return reading
}

func (w *World) enumReading(ent *Entity, ch *model.Channel) (bool, any) {
	if len(ch.EnumValues) == 0 {
		return true, nil
	}
	rng := w.substream(ent.ID + "/" + ch.Name + "/enum")
	return true, ch.EnumValues[rng.Intn(len(ch.EnumValues))]
}

// quantize rounds v to the channel's declared resolution.
func quantize(v, resolution float64) float64 {
	if resolution <= 0 {
		return v
	}
	return math.Round(v/resolution) * resolution
}
