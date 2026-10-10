package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// constWorld builds a world with a constant driving input u and one dynamic
// state x.
func constWorld(t *testing.T, mutate func(*model.DomainSpec), start int64) *World {
	t.Helper()
	spec := testSpec(t, mutate)
	return newTestWorld(t, spec, 42, start)
}

// TestRCNNetworkAnalyticOracle: rc_network is two cascaded first-order lags.
// The closed form of the cascade, from a constant input u and zero initial
// conditions:
//
//	a(t) = gu(1-e^(-t/tau1))
//	x(t) = gu(1-e^(-t/tau2)) - gu(e^(-t/tau2)-e^(-t/tau1))/(1/tau1-1/tau2)
func TestRCNNetworkAnalyticOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	closed := func(t, tau1, tau2, g, u float64) float64 {
		e1 := math.Exp(-t / tau1)
		e2 := math.Exp(-t / tau2)
		den := 1/tau1 - 1/tau2
		return g*u*(1-e2) - g*u*(e2-e1)/(tau2*den)
	}
	for _, tau1 := range []float64{300, 3600} {
		for _, tau2 := range []float64{100, 1200} {
			for _, g := range []float64{0.5, 2} {
				w := constWorld(t, func(s *model.DomainSpec) {
					s.Dynamics[0].F1.Form = "rc_network"
					s.Dynamics[0].F1.TimeConstantS = tau1
					s.Dynamics[0].F1.TimeConstant2S = tau2
					s.Dynamics[0].F1.Gain = g
				}, start)
				u := 2.0 // testSpec state u initial
				horizon := 3 * math.Max(tau1, tau2)
				for i := 1; i <= 20; i++ {
					tS := horizon * float64(i) / 20
					tNS := start + int64(tS*secondsPerNS)
					got := w.StateValue("e-1", "x", tNS)
					want := closed(tS, tau1, tau2, g, u)
					if math.Abs(got-want) > 0.01 {
						t.Fatalf("tau1=%v tau2=%v g=%v t=%v: got %v, closed form %v (Δ=%v)", tau1, tau2, g, tS, got, want, math.Abs(got-want))
					}
				}
			}
		}
	}
}

// TestIntegratorAnalyticOracle: dx/dt = g*u, closed form x(t) = x0 + g*u*t.
func TestIntegratorAnalyticOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	for _, g := range []float64{0.5, 1, 3} {
		w := constWorld(t, func(s *model.DomainSpec) {
			s.Dynamics[0].F1.Form = "integrator"
			s.Dynamics[0].F1.Gain = g
		}, start)
		u := 2.0
		for _, hours := range []float64{0.5, 2, 6} {
			tS := hours * 3600
			got := w.StateValue("e-1", "x", start+int64(tS*secondsPerNS))
			want := g * u * tS
			if math.Abs(got-want) > 1e-6 {
				t.Fatalf("g=%v t=%v: got %v, closed form %v", g, tS, got, want)
			}
		}
	}
}

// TestThresholdIntegratorAnalyticOracle: dx/dt = g*max(0, threshold-u)
// (direction below): linear while u is below the threshold, frozen above.
func TestThresholdIntegratorAnalyticOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	// u = 2 constant; threshold 1.5 -> u above threshold -> frozen.
	w := constWorld(t, func(s *model.DomainSpec) {
		s.Dynamics[0].F1.Form = "threshold_integrator"
		s.Dynamics[0].F1.Gain = 2
		s.Dynamics[0].F1.Threshold = 1.5
		s.Dynamics[0].F1.Direction = "below"
	}, start)
	for _, hours := range []float64{1, 5} {
		tNS := start + int64(hours*3600*secondsPerNS)
		if got := w.StateValue("e-1", "x", tNS); got != 0 {
			t.Fatalf("above-threshold input must freeze the state, got %v", got)
		}
	}
	// Threshold 5 -> u below -> linear growth at g*(5-2)=6 per second.
	w = constWorld(t, func(s *model.DomainSpec) {
		s.Dynamics[0].F1.Form = "threshold_integrator"
		s.Dynamics[0].F1.Gain = 2
		s.Dynamics[0].F1.Threshold = 5
		s.Dynamics[0].F1.Direction = "below"
	}, start)
	tS := 2 * 3600.0
	got := w.StateValue("e-1", "x", start+int64(tS*secondsPerNS))
	if want := 6 * tS; math.Abs(got-want) > 1e-6 {
		t.Fatalf("below-threshold growth: got %v, closed form %v", got, want)
	}
}

// TestDeadTimeDiscreteOracle: dead_time is a ring buffer of
// ceil(dead_time_s/dt) samples feeding a first-order lag. The oracle is an
// independent reimplementation of the exact discrete recurrence, driven by a
// time-varying input (a step fault on the driving state u) so the delay is
// observable.
func TestDeadTimeDiscreteOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	const (
		dtNS      = 1 * secondsPerNS
		deadTimeS = 2.5
		tau       = 4.0
		gain      = 1.0
		faultAtS  = 5.0 // seconds into the run
		horizonS  = 25.0
	)
	uBase := 2.0 // testSpec state u initial
	faultAtNS := start + int64(faultAtS*secondsPerNS)
	horizonNS := start + int64(horizonS*secondsPerNS)
	u := func(t int64) float64 {
		if t >= faultAtNS {
			return uBase + 1 // the step fault adds 1
		}
		return uBase
	}
	// Independent recurrence: same ring semantics (fill until full, then
	// read cap samples back), same forward-Euler lag update.
	cap := int(math.Ceil(deadTimeS / (float64(dtNS) / secondsPerNS)))
	oracleX := func() float64 {
		buf := make([]float64, cap)
		head := 0
		full := false
		x := 0.0
		for t := start; t < horizonNS; t += dtNS {
			var delayed float64
			if !full {
				delayed = u(t)
			} else {
				delayed = buf[head]
			}
			x += (gain*delayed - x) / tau * (float64(dtNS) / secondsPerNS)
			buf[head] = u(t)
			head = (head + 1) % cap
			if head == 0 {
				full = true
			}
		}
		return x
	}
	w := constWorld(t, func(s *model.DomainSpec) {
		s.Dynamics[0].F1.Form = "dead_time"
		s.Dynamics[0].F1.TimeConstantS = tau
		s.Dynamics[0].F1.DeadTimeS = deadTimeS
		s.Dynamics[0].F1.Gain = gain
		s.Dynamics[0].DTMs = 1000 // 1s steps
		// A fault on the driving state u: baseline before the onset, +1 after.
		s.Faults = append(s.Faults, model.Fault{
			ID: "ufault", Onset: model.FaultOnset{Shape: "step"},
			Affects: []model.FaultEffect{{State: "u", Delta: 1}},
			Observability: model.Observability{
				Detector: model.Detector{Form: "single_channel_snr", Channel: "ch"},
			},
		})
	}, start)
	if _, err := w.InjectFault("e-1", "ufault", faultAtNS, nil); err != nil {
		t.Fatal(err)
	}
	want := w.StateValue("e-1", "x", horizonNS)
	if got := oracleX(); math.Abs(got-want) > 1e-9 {
		t.Fatalf("dead_time oracle mismatch: independent recurrence %v, world %v", got, want)
	}
	// The delay must be visible: the response must lag the step at 5s. With
	// no delay the lag state would already exceed 1.5 by t=8s.
	if x := oracleX(); x < 1.5 {
		t.Fatalf("dead_time never responded to the step: x=%v", x)
	}
}

// TestF0TrendSeasonalityOracle: the F0 generator is an analytic sum —
// baseline + trend*h + drift*h + sum amplitude*sin(2*pi*t/period + phase).
func TestF0TrendSeasonalityOracle(t *testing.T) {
	start := model.DefaultStartTimeNS
	w := constWorld(t, func(s *model.DomainSpec) {
		s.Dynamics[0].Tier = "F0"
		s.Dynamics[0].F1 = nil
		s.Dynamics[0].F0 = &model.F0Dyn{
			Baseline:     2.0,
			TrendPerHour: 0.5,
			DriftPerHour: 0.1,
			Seasonality: []model.Seasonality{
				{Amplitude: 0.8, PeriodS: 3600, PhaseRad: 0.5},
				{Amplitude: 0.3, PeriodS: 7200, PhaseRad: 1.2},
			},
		}
	}, start)
	closed := func(tS float64) float64 {
		h := tS / 3600
		v := 2.0 + 0.5*h + 0.1*h
		v += 0.8 * math.Sin(2*math.Pi*tS/3600+0.5)
		v += 0.3 * math.Sin(2*math.Pi*tS/7200+1.2)
		return v
	}
	for _, hours := range []float64{0.25, 1.5, 7} {
		tS := hours * 3600
		got := w.StateValue("e-1", "x", start+int64(tS*secondsPerNS))
		if want := closed(tS); math.Abs(got-want) > 1e-9 {
			t.Fatalf("F0 oracle: got %v, closed form %v", got, want)
		}
	}
}

// TestFaultEnvelopeOracles: each deterministic onset shape has a closed-form
// envelope; the faulted state value must match the full closed form —
// natural lag growth plus the fault's delta*envelope contribution.
func TestFaultEnvelopeOracles(t *testing.T) {
	start := model.DefaultStartTimeNS
	rate := 2.0 // per hour
	for _, shape := range []string{"step", "ramp", "exponential", "intermittent"} {
		tau := 3600.0
		w := constWorld(t, func(s *model.DomainSpec) {
			s.Dynamics[0].F1.TimeConstantS = tau
			s.Faults[0].Onset.Shape = shape
			s.Faults[0].Onset.RatePerHour = rate
		}, start)
		if _, err := w.InjectFault("e-1", "f1", 0, nil); err != nil {
			t.Fatal(err)
		}
		envelope := func(tS float64) float64 {
			switch shape {
			case "step":
				return 1
			case "ramp":
				return math.Min(rate*tS/3600, 1)
			case "exponential":
				return 1 - math.Exp(-rate*tS/3600)
			case "intermittent":
				period := 3600.0 / rate
				cycle := math.Mod(tS, period)
				if cycle < 0.5*period {
					return 1
				}
				return 0
			}
			return 0
		}
		// x(t) = g*u*(1-e^(-t/tau)) + delta*envelope(t) with g=1, u=2 (the
		// constant driving input), x0=0, delta=1.
		closed := func(tS float64) float64 {
			return 2.0*(1-math.Exp(-tS/tau)) + 1.0*envelope(tS)
		}
		for _, atS := range []float64{0.25 * 3600, 1.5 * 3600, 3 * 3600} {
			tNS := start + int64(atS*secondsPerNS)
			got := w.StateValue("e-1", "x", tNS)
			want := closed(atS)
			if math.Abs(got-want) > 0.01 {
				t.Fatalf("envelope %s at t=%v: got %v, closed form %v", shape, atS, got, want)
			}
		}
	}
}
