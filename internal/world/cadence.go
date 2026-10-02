package world

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

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
