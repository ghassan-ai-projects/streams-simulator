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

func (k *kick) assignedFrom(underlying float64, t int64) float64 {
	if k.tauNS <= 0 {
		return k.delta
	}
	return k.delta - (k.delta-underlying)*math.Exp(-float64(t-k.startNS)/k.tauNS)
}

type activeFault struct {
	fault     *model.Fault
	entity    string
	onsetNS   int64
	clearedNS int64
	severity  float64

	walks map[string][]float64
}

func (f *activeFault) envelopeAt(state string, t int64, rng *randutil.SplitMix64, dtNS int64) float64 {
	elapsed := t - f.onsetNS
	if elapsed < 0 {
		return 0
	}
	if f.clearedNS > 0 && t >= f.clearedNS {
		return 0
	}
	return f.onsetEnvelope(state, elapsed, rng, dtNS)
}

func (f *activeFault) contributionAt(state string, t int64, rng *randutil.SplitMix64, dtNS int64) (add, mult float64) {
	for _, a := range f.fault.Affects {
		if a.State != state {
			continue
		}
		e := f.envelopeAt(state, t, rng, dtNS) * f.severity
		if a.Multiplicative {
			mult += a.Delta * e
		} else {
			add += a.Delta * e
		}
	}
	return add, mult
}

func (f *activeFault) onsetEnvelope(state string, elapsed int64, rng *randutil.SplitMix64, dt int64) float64 {
	rate := f.fault.Onset.RatePerHour
	if f.fault.Onset.Shape == "stochastic" {
		return f.stochasticEnvelope(state, elapsed, rng, dt, rate)
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

func (f *activeFault) stochasticEnvelope(state string, elapsed int64, rng *randutil.SplitMix64, dt int64, rate float64) float64 {
	if dt <= 0 {
		dt = 60 * secondsPerNS
	}
	if f.walks == nil {
		f.walks = map[string][]float64{}
	}
	walk := f.walks[state]
	for k := int(elapsed / dt); len(walk) <= k; {
		walk = append(walk, nextWalk(walk, rng, float64(dt)/secondsPerNS/3600*rate))
	}
	f.walks[state] = walk
	return math.Min(math.Max(walk[elapsed/dt], 0), 1)
}

func nextWalk(walk []float64, rng *randutil.SplitMix64, drift float64) float64 {
	if len(walk) == 0 {
		return 0
	}
	return walk[len(walk)-1] + drift + 0.5*math.Sqrt(drift+1e-12)*math.Abs(rng.Norm())
}

func intermittentPeriod(rate float64) int64 {
	period := 3600.0
	if rate > 0 {
		period = 3600 / rate
	}
	return int64(period * secondsPerNS)
}
