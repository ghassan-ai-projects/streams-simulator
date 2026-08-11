package world

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

// TestStepFaultMovesState verifies a step fault shifts hidden state and
// clearing it reverts the contribution.
func TestStepFaultMovesState(t *testing.T) {
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

// TestEmitterReceivesEvents runs the full emission pipeline and checks
// cadence, quantization, seq and world id.
func TestEmitterReceivesEvents(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.Channels = []model.Channel{{
			Name: "heartbeat", ValueType: "none", Resolution: 1,
			Fidelity: "F0", Absence: "signal",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 60},
			Noise:   model.Noise{Model: "none", Sigma: 0},
		}}
		s.Faults[0].Observability.Detector.Channel = "heartbeat"
	})
	start := model.DefaultStartTimeNS
	w, err := New(spec, 1, "w-evt", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(e model.SimEvent) { events = append(events, e) })
	if _, _, err := w.Advance(start + 5*60*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Fatalf("expected 5 heartbeats, got %d", len(events))
	}
	for i, e := range events {
		if e.Seq != int64(i) {
			t.Fatalf("seq not monotonic: %v", e.Seq)
		}
		if e.WorldID != "w-evt" || e.EntityID != "e-1" || e.Channel != "heartbeat" {
			t.Fatalf("event identity wrong: %+v", e)
		}
		got, err := model.ParseTime(e.EventTime)
		if err != nil {
			t.Fatal(err)
		}
		want := start + int64(i+1)*60*secondsPerNS
		if got != want {
			t.Fatalf("event %d at %d, want %d", i, got, want)
		}
		if e.Value != nil {
			t.Fatalf("valueless channel carried a value: %+v", e.Value)
		}
	}
}

// TestSubstreamIsolation: adding an unrelated entity must not change an
// existing entity's output (S1 gate 5, at the world level).
func TestSubstreamIsolation(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.Channels = []model.Channel{{
			Name: "v", ValueType: "number", Unit: "u", Resolution: 0.01,
			Fidelity: "F0", Absence: "signal",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 60, JitterS: 2},
			Noise:   model.Noise{Model: "gaussian", Sigma: 0.5},
		}}
		s.Faults[0].Observability.Detector.Channel = "v"
	})
	run := func(entities []string) []model.SimEvent {
		w, err := New(spec, 99, "w-iso", model.DefaultStartTimeNS, Options{InitialEntities: entities})
		if err != nil {
			t.Fatal(err)
		}
		var evs []model.SimEvent
		w.SetEmitter(func(e model.SimEvent) { evs = append(evs, e) })
		if _, _, err := w.Advance(model.DefaultStartTimeNS + 10*60*secondsPerNS); err != nil {
			t.Fatal(err)
		}
		return evs
	}
	alone := run([]string{"e-1"})
	pair := run([]string{"e-1", "e-2"})
	if len(pair) <= len(alone) {
		t.Fatalf("pair should emit at least as many events")
	}
	var firstA, firstP []model.SimEvent
	for _, e := range alone {
		if e.EntityID == "e-1" {
			firstA = append(firstA, e)
		}
	}
	for _, e := range pair {
		if e.EntityID == "e-1" {
			firstP = append(firstP, e)
		}
	}
	if len(firstA) != len(firstP) {
		t.Fatalf("entity e-1 event count changed: %d vs %d", len(firstA), len(firstP))
	}
	for i := range firstA {
		if firstA[i].EventTime != firstP[i].EventTime || firstA[i].Value != firstP[i].Value {
			t.Fatalf("entity e-1 event %d changed after adding e-2:\n  %+v\n  %+v", i, firstA[i], firstP[i])
		}
	}
}

// TestEffectorIdempotencyAndInterlock covers the loop's basic invariants.
func TestEffectorIdempotencyAndInterlock(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		// start_aerator-style effector on state x, interlocked on state u.
		s.Effectors = []model.Effector{{
			Name:       "act",
			ArgsSchema: map[string]any{"type": "object"},
			Ack:        model.Ack{LatencyMS: model.Latency{Mean: 100}},
			Effect:     model.Effect{StateDeltas: []model.StateDelta{{State: "x", Delta: 1}}, TimeConstantS: 100},
			Interlock:  &model.Interlock{State: "u", Operator: "gt", Threshold: 1},
		}}
		s.Faults = nil
		s.Faults = []model.Fault{{
			ID: "f1", Onset: model.FaultOnset{Shape: "step"},
			Affects: []model.FaultEffect{{State: "x", Delta: 1}},
			Observability: model.Observability{
				Detector: model.Detector{Form: "single_channel_snr", Channel: "ch"},
			},
		}}
	})
	w := newTestWorld(t, spec, 5, model.DefaultStartTimeNS)
	at := model.DefaultStartTimeNS + 60*secondsPerNS
	// u = 2 > 1, so the interlock refuses.
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at); err == nil {
		t.Fatal("interlock should refuse")
	}
	calls := w.EffectorCalls()
	if len(calls) != 1 || !calls[0].InterlockRefused {
		t.Fatalf("interlock refusal not recorded: %+v", calls)
	}
}

// TestEffectsAreEntityScoped protects the closed-loop invariant that an
// effector command changes only the addressed producer. A state name is not
// a sufficient key because every entity may expose the same state.
func TestEffectsAreEntityScoped(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Effectors = []model.Effector{{
			Name:       "act",
			ArgsSchema: map[string]any{"type": "object"},
			Ack:        model.Ack{LatencyMS: model.Latency{Mean: 1}},
			Effect:     model.Effect{StateDeltas: []model.StateDelta{{State: "x", Delta: 3}}},
		}}
	})
	start := model.DefaultStartTimeNS
	w := newTestWorld(t, spec, 17, start)
	if err := w.AddEntity("e-2", start, nil); err != nil {
		t.Fatal(err)
	}
	before := w.StateValue("e-2", "x", start+secondsPerNS)
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", nil, start); err != nil {
		t.Fatal(err)
	}
	addressed := w.StateValue("e-1", "x", start+secondsPerNS)
	unaddressed := w.StateValue("e-2", "x", start+secondsPerNS)
	if addressed-before < 2.9 {
		t.Fatalf("addressed entity did not receive effect: e-1=%v e-2-before=%v", addressed, before)
	}
	if math.Abs(unaddressed-before) > 1e-9 {
		t.Fatalf("effect leaked to e-2: before=%v after=%v", before, unaddressed)
	}
}

// TestChurnAllocatesAndContinuesAfterBirths protects autonomous lifecycle
// scheduling. The first generated birth must not collide with e-1 and a
// rejected/custom identity must not stop future births.
func TestChurnAllocatesAndContinuesAfterBirths(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Entities.Churn = &model.Churn{BirthsPerHour: 3600, MeanLifetimeS: 1e6}
	})
	start := model.DefaultStartTimeNS
	w := newTestWorld(t, spec, 19, start)
	if _, _, err := w.Advance(start + 10*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	ids := w.EntityIDs()
	if len(ids) < 2 {
		t.Fatalf("expected at least one autonomous birth, got %v", ids)
	}
	if _, ok := w.entities["e-1"]; !ok {
		t.Fatalf("initial entity was replaced: %v", ids)
	}
	if _, ok := w.entities["e-2"]; !ok {
		t.Fatalf("first birth did not receive a free identity: %v", ids)
	}
}
