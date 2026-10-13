package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

const secondsPerNS = 1e9

type kick struct {
	entity  string
	state   string
	delta   float64
	assign  bool
	startNS int64
	tauNS   float64
	applied bool
}

type driverKey struct {
	entity string
	state  string
}

func (k *kick) valueAt(t int64) float64 {
	if t <= k.startNS {
		return 0
	}
	if k.tauNS <= 0 {
		return k.delta
	}
	return k.delta * (1 - math.Exp(-float64(t-k.startNS)/k.tauNS))
}

type activeFault struct {
	fault     *model.Fault
	entity    string
	onsetNS   int64
	clearedNS int64
	severity  float64

	walk     float64
	walkStep int64
}

func (f *activeFault) envelopeAt(t int64, rng *randutil.SplitMix64, dtNS int64) float64 {
	elapsed := t - f.onsetNS
	if elapsed < 0 {
		return 0
	}
	if f.clearedNS > 0 && t >= f.clearedNS {
		return 0
	}
	return f.onsetEnvelope(elapsed, t, rng, dtNS)
}

func (f *activeFault) contributionAt(state string, t int64, rng *randutil.SplitMix64, dtNS int64) (add, mult float64) {
	for _, a := range f.fault.Affects {
		if a.State != state {
			continue
		}
		e := f.envelopeAt(t, rng, dtNS) * f.severity
		if a.Multiplicative {
			mult += a.Delta * e
		} else {
			add += a.Delta * e
		}
	}
	return add, mult
}

func (f *activeFault) onsetEnvelope(elapsed, at int64, rng *randutil.SplitMix64, dt int64) float64 {
	rate := f.fault.Onset.RatePerHour
	if f.fault.Onset.Shape == "stochastic" {
		return f.stochasticEnvelope(at, rng, dt, rate)
	}
	return f.deterministicEnvelope(elapsed, rate)
}

func (f *activeFault) deterministicEnvelope(elapsed int64, rate float64) float64 {
	switch f.fault.Onset.Shape {
	case "step", model.ShapePulse:
		return 1
	case "ramp":
		return math.Min(rate*(float64(elapsed)/secondsPerNS/3600), 1)
	case "exponential":
		return 1 - math.Exp(-rate*float64(elapsed)/secondsPerNS/3600)
	case "intermittent":
		return f.intermittentEnvelope(elapsed, rate)
	}
	return 0
}

func (f *activeFault) intermittentEnvelope(elapsed int64, rate float64) float64 {
	periodNS := intermittentPeriod(rate)
	cycle := elapsed % periodNS
	duty := f.fault.Onset.DutyCycle
	if duty <= 0 {
		duty = 0.5
	}
	if float64(cycle) < duty*float64(periodNS) {
		return 1
	}
	return 0
}

func (f *activeFault) stochasticEnvelope(at int64, rng *randutil.SplitMix64, dt int64, rate float64) float64 {
	if dt <= 0 {
		dt = 60 * secondsPerNS
	}

	for f.walkStep+dt <= at {
		next := f.walkStep + dt
		hours := float64(next-f.walkStep) / secondsPerNS / 3600
		f.walk += rate*hours + 0.5*math.Sqrt(rate*hours+1e-12)*math.Abs(rng.Norm())
		f.walkStep = next
	}
	return math.Min(math.Max(f.walk, 0), 1)
}

func intermittentPeriod(rate float64) int64 {
	period := 3600.0
	if rate > 0 {
		period = 3600 / rate
	}
	return int64(period * secondsPerNS)
}
