package world

// The emission pipeline: at each scheduled channel emission, compute the
// true state, apply the channel's observation function (gain, offset,
// observation bias, noise, drift), quantize to the channel's resolution,
// apply the link delay, and emit the native event. Then schedule the next
// emission per the cadence model.

import (
	"math"

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

	// Link delay: observed_time = event_time + delay.
	delay := w.linkDelay(entityID, ch, t)
	observed := t + int64(delay*secondsPerNS)

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

	// Schedule the next emission for this channel.
	next := w.nextEmission(entityID, ch, t)
	if next > 0 {
		w.schedule(kindEmission, entityID, channelName, next, nil)
	}
}

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

// cadenceEmits decides whether this tick produces a record.
func (w *World) cadenceEmits(ch *model.Channel, cs *channelRunState, t int64, raw float64) bool {
	switch ch.Cadence.Mode {
	case "periodic":
		return true
	case "report_by_exception":
		if !cs.hasSent {
			cs.lastSent = raw
			cs.hasSent = true
			return true
		}
		return math.Abs(raw-cs.lastSent) >= ch.Cadence.Deadband
	case "event_driven":
		// Numeric event-driven channels emit on state change beyond
		// resolution (string handling is in reading).
		if ch.ValueType == "string" {
			return false
		}
		cur := raw
		changed := !cs.hasSent || math.Abs(cur-cs.lastTrigger) >= ch.Resolution
		cs.lastTrigger = cur
		return changed
	case "batch":
		return true
	case "human_driven":
		return true
	}
	return true
}

// nextEmission computes the next scheduled emission time for a channel.
func (w *World) nextEmission(entityID string, ch *model.Channel, t int64) int64 {
	cd := ch.Cadence
	switch cd.Mode {
	case "periodic", "batch":
		if cd.PeriodS <= 0 {
			return 0
		}
		periodNS := int64(cd.PeriodS * secondsPerNS)
		next := t + periodNS
		if cd.JitterS > 0 && !w.Noiseless {
			rng := w.substream(entityID + "/" + ch.Name + "/jitter")
			j := (2*rng.Float64() - 1) * cd.JitterS
			next += int64(j * secondsPerNS)
		}
		return next
	case "report_by_exception":
		check := int64(60 * secondsPerNS)
		if cd.PeriodS > 0 {
			check = int64(cd.PeriodS * secondsPerNS)
		}
		return t + check
	case "event_driven":
		check := int64(60 * secondsPerNS)
		if cd.PeriodS > 0 {
			check = int64(cd.PeriodS * secondsPerNS)
		}
		return t + check
	case "human_driven":
		if cd.PeriodS <= 0 {
			return 0
		}
		rng := w.substream(entityID + "/" + ch.Name + "/human")
		return t + int64(rng.Exp(cd.PeriodS)*secondsPerNS)
	}
	return 0
}

// linkDelay samples the channel's transport delay.
func (w *World) linkDelay(entityID string, ch *model.Channel, t int64) float64 {
	ld := ch.LinkDelay
	if ld == nil || ld.Model == "" {
		return 0
	}
	rng := w.substream("link/" + entityID + "/" + ch.Name + "/delay")
	switch ld.Model {
	case "constant":
		return ld.MeanS
	case "exponential":
		d := rng.Exp(ld.MeanS)
		if ld.MaxS > 0 && d > ld.MaxS {
			d = ld.MaxS
		}
		return d
	case "lognormal":
		sigma := ld.SigmaS
		if sigma <= 0 {
			sigma = ld.MeanS / 2
		}
		mu := math.Log(ld.MeanS) - sigma*sigma/2
		return rng.Lognormal(mu, sigma)
	case "store_and_forward":
		// A gap that delivers everything late: mean plus an exponential tail.
		return ld.MeanS + rng.Exp(ld.MeanS)
	}
	return 0
}

// updateAvailability advances the producer up/down renewal process.
func (w *World) updateAvailability(ent *Entity, ch *model.Channel, cs *channelRunState, t int64) error {
	rng := w.substream(ent.ID + "/" + ch.Name + "/availability")
	a := ch.Availability
	if a.Uptime >= 1 {
		cs.availInit = true
		cs.availDown = false
		cs.availUntil = 0
		return nil
	}
	if !cs.availInit {
		cs.availInit = true
		cs.availDown = false
		cs.availUntil = t + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS)))
		return nil
	}
	if t < cs.availUntil {
		return nil
	}
	if cs.availDown {
		cs.availDown = false
		cs.availUntil = t + int64(rng.Exp(mtbfSeconds(a.Uptime, a.MTTRS)))
	} else {
		cs.availDown = true
		cs.availUntil = t + int64(rng.Exp(mttrSeconds(a.MTTRS)))
	}
	return nil
}

func mttrSeconds(mttr float64) float64 {
	if mttr <= 0 {
		return 1
	}
	return mttr
}

func mtbfSeconds(uptime, mttr float64) float64 {
	if uptime >= 1 {
		return math.Inf(1)
	}
	if uptime <= 0 {
		return mttrSeconds(mttr)
	}
	return mttrSeconds(mttr) * uptime / (1 - uptime)
}

// isConfirmationChannel reports whether the channel independently reports an
// effector's physical effect (the confirmation_channels declarations).
func (w *World) isConfirmationChannel(name string) bool {
	for i := range w.Spec.Spec.Effectors {
		for _, c := range w.Spec.Spec.Effectors[i].ConfirmationChannels {
			if c == name {
				return true
			}
		}
	}
	return false
}

// birthAutonomous creates an entity on the churn schedule.
func (w *World) birthAutonomous(atNS int64) {
	ch := w.Spec.Spec.Entities.Churn
	if ch == nil {
		return
	}
	w.nextIndex++
	id := renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	for id == "" || w.entities[id] != nil {
		w.nextIndex++
		id = renderID(w.Spec.Spec.Entities.IDTemplate, w.nextIndex, nil)
	}
	if err := w.addEntity(id, atNS, nil); err == nil {
		// Schedule its death.
		if ch.MeanLifetimeS > 0 {
			rng := w.substream("churn/lifetimes/" + id)
			dieAt := atNS + int64(rng.Exp(ch.MeanLifetimeS)*secondsPerNS)
			w.schedule(kindDeath, id, "", dieAt, nil)
		}
	}
	// Schedule the next birth.
	rng := w.substream("churn/births")
	next := atNS + int64(rng.Exp(3600*secondsPerNS/ch.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, nil)
}

// Retire removes an entity and its scheduled emissions. Used for churn and
// by the run layer for entity.retire commands.
func (w *World) Retire(entityID, reason string, atNS int64) {
	w.retire(entityID, reason, atNS)
}

func (w *World) retire(entityID, reason string, atNS int64) {
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return
	}
	ent.alive = false
	ent.RetiredNS = atNS
	// Scheduled emissions are dropped when popped (processEmission checks
	// liveness).
}
