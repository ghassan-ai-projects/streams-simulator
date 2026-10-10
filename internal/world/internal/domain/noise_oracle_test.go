package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// TestGaussianNoiseMomentsOracle: gaussian noise on a quiet channel has the
// declared sigma as its empirical standard deviation (deterministic seed).
func TestGaussianNoiseMomentsOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.Channels = []model.Channel{{
			Name: "noisy", ValueType: "number", Unit: "u", Resolution: 1e-9,
			Fidelity: "F0", Absence: "signal", Observes: "x",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 0.01},
			Noise:   model.Noise{Model: "gaussian", Sigma: 0.25},
		}}
		s.Faults[0].Observability.Detector.Channel = "noisy"
	})
	w, err := New(spec, 5, "w-noise", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var values []float64
	w.SetEmitter(func(ev model.SimEvent) {
		if v, ok := asFloatTest(ev.Value); ok {
			values = append(values, v)
		}
	})
	if _, _, err := w.Advance(start + int64(3600*secondsPerNS)); err != nil {
		t.Fatal(err)
	}
	if len(values) < 1000 {
		t.Fatalf("need many samples for the moment oracle, got %d", len(values))
	}
	mean, std := meanStdOracle(values)
	if math.Abs(std-0.25) > 0.05 {
		t.Fatalf("gaussian oracle: empirical std %v, declared sigma 0.25", std)
	}
	if math.Abs(mean) > 0.05 {
		t.Fatalf("gaussian oracle: mean must be ~0, got %v", mean)
	}
}

// TestQuantizationNoiseOracle: quantization noise forces every emitted value
// onto the sigma grid — no value may fall off it.
func TestQuantizationNoiseOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.State[0].Initial = 2.25
		s.Channels = []model.Channel{{
			Name: "quant", ValueType: "number", Unit: "u", Resolution: 1e-9,
			Fidelity: "F0", Absence: "signal", Observes: "x",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 0.01},
			Noise:   model.Noise{Model: "quantization", Sigma: 0.5},
		}}
		s.Faults[0].Observability.Detector.Channel = "quant"
	})
	w, err := New(spec, 5, "w-quant", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	w.SetEmitter(func(ev model.SimEvent) {
		v, ok := asFloatTest(ev.Value)
		if !ok {
			return
		}
		q := math.Round(v / 0.5)
		if math.Abs(v-q*0.5) > 1e-9 {
			t.Fatalf("quantization oracle: value %v is off the 0.5 grid", v)
		}
	})
	if _, _, err := w.Advance(start + int64(600*secondsPerNS)); err != nil {
		t.Fatal(err)
	}
}

func asFloatTest(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	}
	return 0, false
}

func meanStdOracle(xs []float64) (float64, float64) {
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(sq / float64(len(xs)))
}

// TestAvailabilitySojournOracle: availability is an exponential up/down
// renewal process with mtbf = mttr*uptime/(1-uptime). Over a long horizon
// the empirical duty cycle must approach the declared uptime, and the
// producer must be silent during down periods (observed as gaps several
// channel periods long).
func TestAvailabilitySojournOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.Channels = []model.Channel{{
			Name: "avail", ValueType: "number", Unit: "u", Resolution: 1e-9,
			Fidelity: "F0", Absence: "signal", Observes: "x",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 30},
			Noise:   model.Noise{Model: "none", Sigma: 0},
			Availability: &model.Availability{
				Uptime: 0.5, MTTRS: 60,
			},
		}}
		s.Faults[0].Observability.Detector.Channel = "avail"
	})
	w, err := New(spec, 5, "w-avail", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var emitted []int64
	w.SetEmitter(func(ev model.SimEvent) {
		if t, err := model.ParseTime(ev.ObservedTime); err == nil {
			emitted = append(emitted, t)
		}
	})
	horizon := int64(24 * 3600 * secondsPerNS)
	if _, _, err := w.Advance(start + horizon); err != nil {
		t.Fatal(err)
	}
	if len(emitted) == 0 {
		t.Fatal("availability must not silence the producer forever")
	}
	// Silence windows: with a 30s cadence and ~60s down sojourns, gaps of
	// several periods must appear.
	maxGap := int64(0)
	var downNS int64
	for i := 1; i < len(emitted); i++ {
		g := emitted[i] - emitted[i-1]
		if g > maxGap {
			maxGap = g
		}
		// A gap beyond one and a half cadence periods is a down sojourn.
		if g > 3*30*secondsPerNS/2 {
			downNS += g
		}
	}
	if maxGap < 2*30*secondsPerNS {
		t.Fatalf("no down period observed: max inter-emission gap %v", maxGap)
	}
	// Duty cycle from the gaps: up time is the horizon minus the down
	// sojourns, an unbiased estimate (bucketing by fixed windows is not).
	duty := 1 - float64(downNS)/float64(horizon)
	if math.Abs(duty-0.5) > 0.15 {
		t.Fatalf("availability duty cycle %v does not approach the declared uptime 0.5", duty)
	}
}
