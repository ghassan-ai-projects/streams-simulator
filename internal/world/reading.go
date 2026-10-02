package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

// reading computes the observed (pre-quantization) value and whether this
// tick emits, per the channel's cadence mode.
func (w *World) reading(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) (bool, any) {
	var raw float64
	switch ch.ValueType {
	case "none":
		return w.cadenceEmits(ch, cs, t, 0), nil
	case "string":
		// Event-driven string channels emit on trigger-state change; the
		// value is sampled from the declared enum values.
		if ch.Cadence.Mode == "event_driven" {
			trig := ch.Cadence.TriggerState
			cur := w.stateAt(ent.ID, trig, t)
			changed := !cs.hasSent || math.Abs(cur-cs.lastTrigger) >= ch.Resolution
			cs.lastTrigger = cur
			if !changed {
				return false, nil
			}
			cs.hasSent = true
			if len(ch.EnumValues) == 0 {
				return true, nil
			}
			rng := w.substream(ent.ID + "/" + ch.Name + "/enum")
			idx := rng.Intn(len(ch.EnumValues))
			return true, ch.EnumValues[idx]
		}
		// Non-event-driven string channels emit on the base cadence.
		return w.cadenceEmits(ch, cs, t, 0), nil
	default:
		raw = w.observe(ent, ch, cs, t)
	}
	emit := w.cadenceEmits(ch, cs, t, raw)
	if !emit {
		return false, nil
	}
	cs.lastSent = raw
	cs.hasSent = true
	switch ch.ValueType {
	case "boolean":
		return true, raw >= 0.5
	case "counter":
		return true, int64(math.Round(raw))
	default:
		return true, raw
	}
}

// observe applies the channel's observation function to the hidden state.
func (w *World) observe(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) float64 {
	state := ch.Observes
	v := w.stateAt(ent.ID, state, t)
	// A confirmation channel under an active silent_no_effect shadow reports
	// the counterfactual: add the shadow kicks.
	if w.isConfirmationChannel(ch.Name) {
		v = w.shadowValue(ent.ID, state, t)
	}
	gain := w.Spec.ChannelGain(ch.Name)
	reading := gain*v + ch.ObservationOffset
	if b := ch.ObservationBias; b != nil {
		bias := w.stateAt(ent.ID, b.State, t)
		reading += b.Coef * bias
	}
	// Observation noise, scaled by the bias level toward noise_scale_at_full.
	sigma := ch.Noise.Sigma
	if b := ch.ObservationBias; b != nil && sigma > 0 && !w.Noiseless {
		bias := math.Min(math.Max(w.stateAt(ent.ID, b.State, t), 0), 1)
		scale := 1 + (b.NoiseScaleAtFull-1)*bias
		sigma *= scale
	}
	if sigma > 0 && ch.Noise.Model == "quantization" && !w.Noiseless {
		reading = math.Round(reading/sigma) * sigma
	} else if sigma > 0 && !w.Noiseless {
		rng := w.substream(ent.ID + "/" + ch.Name + "/noise")
		reading += sigma * rng.Norm()
	}
	// Drift on the reading.
	if d := ch.Drift; d != nil && d.RatePerHour != 0 {
		reading += w.drift(ent, ch, cs, t, d)
	}
	// Quantize to the channel resolution.
	return quantize(reading, ch.Resolution)
}

// drift advances the reading drift (linear or random walk).
func (w *World) drift(ent *Entity, ch *model.Channel, cs *channelRunState, t int64, d *model.Drift) float64 {
	if cs.walkLastNS == 0 {
		cs.walkLastNS = t
		return d.RatePerHour * (float64(t-w.StartNS) / secondsPerNS / 3600)
	}
	dtH := float64(t-cs.walkLastNS) / secondsPerNS / 3600
	cs.walkLastNS = t
	switch d.Model {
	case "random_walk":
		rng := w.substream(ent.ID + "/" + ch.Name + "/drift")
		step := d.RatePerHour*dtH + math.Abs(d.RatePerHour)*math.Sqrt(dtH)*rng.Norm()
		cs.walk += step
		return cs.walk
	default: // linear
		return d.RatePerHour * (float64(t-w.StartNS) / secondsPerNS / 3600)
	}
}
