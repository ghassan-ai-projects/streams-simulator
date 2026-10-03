package world

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// cadenceEmits decides whether this tick produces a record.
func (w *World) cadenceEmits(ch *model.Channel, cs *channelRunState, t int64, raw float64) bool {
	switch ch.Cadence.Mode {
	case "report_by_exception":
		return exceptionCadence(ch, cs, raw)
	case "event_driven":
		return eventCadence(ch, cs, raw)
	}
	return true
}

// nextEmission computes the next scheduled emission time for a channel.
func (w *World) nextEmission(entityID string, ch *model.Channel, t int64) int64 {
	switch ch.Cadence.Mode {
	case "periodic", "batch":
		return w.periodicEmission(entityID, ch, t)
	case "report_by_exception", "event_driven":
		return checkedEmission(ch, t)
	case "human_driven":
		return w.humanEmission(entityID, ch, t)
	}
	return 0
}

// linkDelay samples the channel's transport delay.
func (w *World) linkDelay(entityID string, ch *model.Channel, t int64) float64 {
	delay := ch.LinkDelay
	if delay == nil || delay.Model == "" {
		return 0
	}
	rng := w.substream("link/" + entityID + "/" + ch.Name + "/delay")
	return sampleLinkDelay(delay, rng)
}

func exceptionCadence(ch *model.Channel, cs *channelRunState, raw float64) bool {
	if !cs.hasSent {
		cs.lastSent = raw
		cs.hasSent = true
		return true
	}
	return math.Abs(raw-cs.lastSent) >= ch.Cadence.Deadband
}

func eventCadence(ch *model.Channel, cs *channelRunState, raw float64) bool {
	if ch.ValueType == "string" {
		return false
	}
	changed := !cs.hasSent || math.Abs(raw-cs.lastTrigger) >= ch.Resolution
	cs.lastTrigger = raw
	return changed
}

func (w *World) periodicEmission(entity string, ch *model.Channel, at int64) int64 {
	cadence := ch.Cadence
	if cadence.PeriodS <= 0 {
		return 0
	}
	next := at + int64(cadence.PeriodS*secondsPerNS)
	if cadence.JitterS > 0 && !w.Noiseless {
		rng := w.substream(entity + "/" + ch.Name + "/jitter")
		jitter := (2*rng.Float64() - 1) * cadence.JitterS
		next += int64(jitter * secondsPerNS)
	}
	return next
}

func checkedEmission(ch *model.Channel, at int64) int64 {
	check := int64(60 * secondsPerNS)
	if ch.Cadence.PeriodS > 0 {
		check = int64(ch.Cadence.PeriodS * secondsPerNS)
	}
	return at + check
}

func (w *World) humanEmission(entity string, ch *model.Channel, at int64) int64 {
	if ch.Cadence.PeriodS <= 0 {
		return 0
	}
	rng := w.substream(entity + "/" + ch.Name + "/human")
	return at + int64(rng.Exp(ch.Cadence.PeriodS)*secondsPerNS)
}
