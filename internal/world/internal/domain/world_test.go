package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// testSpec builds a tiny valid domain spec for world tests.
func testSpec(t *testing.T, mutate func(*model.DomainSpec)) *domain.Compiled {
	t.Helper()
	spec := &model.DomainSpec{
		ID:       "test-world",
		Version:  "0.1.0",
		Title:    "test world",
		Stresses: strings.Repeat("z", 40),
		Axes: model.Axes{
			Rate: "low", Cardinality: "singleton", ValueShape: []string{"scalar"},
			Cadence: []string{"periodic"}, Lateness: "none", Absence: "signal",
			TimeRef: "wall", Correlation: []string{"independent"},
			Seasonality: []string{"none"}, Actuation: "observe_only",
			Consequence: []string{"cost"}, Fidelity: []string{"F0"},
		},
		Entities: model.Entities{IDTemplate: "e-{n}", Count: struct {
			Default int `json:"default"`
			Min     int `json:"min,omitempty"`
			Max     int `json:"max,omitempty"`
		}{Default: 1}},
		State: []model.State{
			{Name: "u", Initial: 2},
			{Name: "x", Initial: 0},
		},
		Dynamics: []model.Dynamics{{
			Target: "x",
			Tier:   "F1",
			DTMs:   1000,
			F1: &model.F1Dyn{
				Form:          "first_order_lag",
				TimeConstantS: 3600,
				Gain:          1,
				Inputs:        []model.F1Input{{State: "u", Coef: 1}},
			},
		}},
		Channels: []model.Channel{{
			Name: "ch", ValueType: "number", Unit: "u", Resolution: 0.01,
			Fidelity: "F1", Absence: "signal", Observes: "x",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 60},
			Noise:   model.Noise{Model: "gaussian", Sigma: 0.1},
		}},
		Faults: []model.Fault{{
			ID: "f1", Onset: model.FaultOnset{Shape: "step"},
			Affects: []model.FaultEffect{{State: "x", Delta: 1}},
			Observability: model.Observability{
				Detector: model.Detector{Form: "single_channel_snr", Channel: "ch"},
			},
		}},
		Profiles: []model.Profile{
			{Name: "nominal", Description: "x"},
			{Name: "correlated_cascade", Description: "x", NotApplicable: strings.Repeat("y", 20)},
			{Name: "sensor_pathology", Description: "x"},
		},
		GroundTruth: model.GroundTruth{NegativeClassFraction: 0.4},
	}
	if mutate != nil {
		mutate(spec)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	c, err := domain.Parse(raw, "test")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// newTestWorld builds a world from the test spec, suppressing emission.
func newTestWorld(t *testing.T, spec *domain.Compiled, seed uint64, startNS int64) *World {
	t.Helper()
	w, err := New(spec, seed, "w-test", startNS, Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TestAnalyticCrossCheck is S1 gate 1: the RK4 integrator must agree with the
// closed-form solution of a first-order lag across a swept parameter space,
// to within the channel resolution. This is the difference between
// "self-consistent" and "right".
func TestAnalyticCrossCheck(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS

	taus := []float64{300, 1200, 3600, 7200}
	gains := []float64{0.5, 1, 2.5}
	us := []float64{1, 3}
	dts := []int64{100, 1000, 5000}
	horizons := []float64{0.25, 1, 3} // in tau multiples

	// closed form: x(t) = g*u + (x0 - g*u)*exp(-t/tau)
	closed := func(t, tau, g, u, x0 float64) float64 {
		return g*u + (x0-g*u)*math.Exp(-t/tau)
	}

	for _, tau := range taus {
		for _, g := range gains {
			for _, u := range us {
				for _, dt := range dts {
					for _, h := range horizons {
						tau := tau
						g := g
						u := u
						dt := dt
						h := h
						t.Run(fmt.Sprintf("tau=%v/g=%v/u=%v/dt=%d/h=%v", tau, g, u, dt, h), func(t *testing.T) {
							t.Parallel()
							spec := testSpec(t, func(s *model.DomainSpec) {
								s.Dynamics[0].DTMs = dt
								s.Dynamics[0].F1.TimeConstantS = tau
								s.Dynamics[0].F1.Gain = g
								s.State[0].Initial = u // driving input
							})
							w := newTestWorld(t, spec, 42, start)
							horizon := tau * h
							steps := 20
							for i := 1; i <= steps; i++ {
								tNS := start + int64(horizon*float64(i)/float64(steps)*secondsPerNS)
								got := w.StateValue("e-1", "x", tNS)
								want := closed(horizon*float64(i)/float64(steps), tau, g, u, 0)
								if math.Abs(got-want) > 0.005 { // well inside resolution 0.01
									t.Fatalf("t=%v: got %v, closed form %v (Δ=%v)", i, got, want, math.Abs(got-want))
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestRCNetworkUsesBothTimeConstants(t *testing.T) {
	t.Parallel()
	makeWorld := func(tau2 float64) *World {
		spec := testSpec(t, func(s *model.DomainSpec) {
			s.Dynamics[0].F1.Form = "rc_network"
			s.Dynamics[0].F1.TimeConstantS = 10
			s.Dynamics[0].F1.TimeConstant2S = tau2
		})
		return newTestWorld(t, spec, 41, model.DefaultStartTimeNS)
	}
	start := model.DefaultStartTimeNS
	fastSecond := makeWorld(1)
	slowSecond := makeWorld(100)
	at := start + 100*secondsPerNS
	fast := fastSecond.StateValue("e-1", "x", at)
	slow := slowSecond.StateValue("e-1", "x", at)
	if fast-slow < 0.2 {
		t.Fatalf("rc_network second time constant has no effect: fast=%v slow=%v", fast, slow)
	}
}

// TestStepFaultMovesState verifies a step fault shifts hidden state and
// clearing it reverts the contribution.
func TestStepFaultMovesState(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, nil)
	w := newTestWorld(t, spec, 7, model.DefaultStartTimeNS)
	w2 := newTestWorld(t, spec, 7, model.DefaultStartTimeNS) // no-fault control
	base := w.StateValue("e-1", "x", model.DefaultStartTimeNS+3600*secondsPerNS)
	if _, err := w.InjectFault("e-1", "f1", 0, nil); err != nil {
		t.Fatal(err)
	}
	after := w.StateValue("e-1", "x", model.DefaultStartTimeNS+3600*secondsPerNS)
	if after-base < 0.9 {
		t.Fatalf("step fault should add ~1.0, got Δ=%v", after-base)
	}
	faults := w.ListFaults()
	if len(faults) != 1 || faults[0].Fault != "f1" {
		t.Fatalf("fault list wrong: %+v", faults)
	}
	if err := w.ClearFault(faults[0].FaultID, 0); err != nil {
		t.Fatal(err)
	}
	at := model.DefaultStartTimeNS + 7200*secondsPerNS
	reverted := w.StateValue("e-1", "x", at)
	baseAt := w2.StateValue("e-1", "x", at)
	if math.Abs(reverted-baseAt) > 0.01 {
		t.Fatalf("clear should revert contribution, got %v vs no-fault %v", reverted, baseAt)
	}
}
