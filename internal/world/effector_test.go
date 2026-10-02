package world

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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

func TestFaultOnsetMagnitudeScalesDeclaredDelta(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Faults[0].Onset.Magnitude = 2
		s.Faults[0].Affects[0].Delta = 1
	})
	start := model.DefaultStartTimeNS
	w := newTestWorld(t, spec, 23, start)
	base := w.StateValue("e-1", "x", start+secondsPerNS)
	if _, err := w.InjectFault("e-1", "f1", start, nil); err != nil {
		t.Fatal(err)
	}
	got := w.StateValue("e-1", "x", start+secondsPerNS)
	if math.Abs((got-base)-2) > 0.01 {
		t.Fatalf("fault magnitude was not applied: base=%v got=%v", base, got)
	}
}

// TestInjectFaultRejectsUnknownParams: fault parameters are fail-closed; the
// only declared key is severity, and it must be numeric.
func TestInjectFaultRejectsUnknownParams(t *testing.T) {
	spec := testSpec(t, nil)
	w := newTestWorld(t, spec, 23, model.DefaultStartTimeNS)
	if _, err := w.InjectFault("e-1", "f1", 0, map[string]any{"bogus": 1}); err == nil {
		t.Fatal("unknown fault param key must be rejected")
	}
	if _, err := w.InjectFault("e-1", "f1", 0, map[string]any{"severity": "high"}); err == nil {
		t.Fatal("non-numeric severity must be rejected")
	}
	if _, err := w.InjectFault("e-1", "f1", 0, map[string]any{"severity": -1.0}); err == nil {
		t.Fatal("negative severity must be rejected")
	}
	if _, err := w.InjectFault("e-1", "f1", 0, map[string]any{"severity": 2.5}); err != nil {
		t.Fatalf("declared severity rejected: %v", err)
	}
}

func TestDeclaredValueTypesReachNativeEvents(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.State[1].Initial = 1
		s.Channels = []model.Channel{
			{Name: "flag", ValueType: "boolean", Resolution: 1, Fidelity: "F0", Absence: "signal", Observes: "x", Cadence: model.Cadence{Mode: "periodic", PeriodS: 60}, Noise: model.Noise{Model: "none"}},
			{Name: "count", ValueType: "counter", Unit: "items", Resolution: 1, Fidelity: "F0", Absence: "signal", Observes: "x", Cadence: model.Cadence{Mode: "periodic", PeriodS: 60}, Noise: model.Noise{Model: "none"}},
		}
		s.Faults[0].Observability.Detector.Channel = "flag"
	})
	start := model.DefaultStartTimeNS
	w, err := New(spec, 29, "w-types", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(ev model.SimEvent) { events = append(events, ev) })
	if _, _, err := w.Advance(start + 60*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	seen := map[string]any{}
	for _, ev := range events {
		seen[ev.Channel] = ev.Value
	}
	if _, ok := seen["flag"].(bool); !ok {
		t.Fatalf("boolean channel emitted %T: %#v", seen["flag"], seen["flag"])
	}
	if _, ok := seen["count"].(int64); !ok {
		t.Fatalf("counter channel emitted %T: %#v", seen["count"], seen["count"])
	}
}

func TestAvailabilityStartsUpAndUsesRenewalTransitions(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics = nil
		s.Channels[0].Availability = &model.Availability{Uptime: 0.5, MTTRS: 60}
	})
	start := model.DefaultStartTimeNS
	w, err := New(spec, 31, "w-availability", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(ev model.SimEvent) { events = append(events, ev) })
	if _, _, err := w.Advance(start + 10*3600*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("availability renewal incorrectly started the producer down")
	}
}
