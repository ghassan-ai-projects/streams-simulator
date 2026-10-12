package domain

import (
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func sagSpec(t *testing.T, withEventChannel bool) func(*model.DomainSpec) {
	t.Helper()
	return func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "voltage", Initial: 100}, model.State{Name: "resets", Initial: 0}, model.State{Name: "sag_trip", Initial: 0})
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{Target: "sag_trip", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormTrip,
			When: &model.Condition{State: "voltage", Operator: "lt", Threshold: 85}, ForS: 0.2, Latch: true, ResetOn: "resets", TrippedValue: 1}})
		pulse := func(id string, seconds float64) model.Fault {
			return model.Fault{ID: id, Onset: model.FaultOnset{Shape: model.ShapePulse, DurationS: seconds}, Affects: []model.FaultEffect{{State: "voltage", Delta: -20}},
				Observability: model.Observability{Detector: model.Detector{Form: "single_channel_snr", Channel: "ch"}}}
		}
		kind.Faults = append(kind.Faults, pulse("long_sag", 0.3), pulse("short_sag", 0.1))
		if withEventChannel {
			kind.Channels = append(kind.Channels, model.Channel{Name: "pq", ValueType: "number", Unit: "%", Resolution: 0.1, Fidelity: "F1",
				Absence: "signal", Observes: "voltage", Cadence: model.Cadence{Mode: "event_driven", PeriodS: 60}, Noise: model.Noise{Model: "gaussian", Sigma: 0.01}},
				model.Channel{Name: "noisy", ValueType: "number", Unit: "u", Resolution: 0.01, Fidelity: "F1",
					Absence: "signal", Observes: "x", Cadence: model.Cadence{Mode: "event_driven", PeriodS: 60}, Noise: model.Noise{Model: "gaussian", Sigma: 0.1}})
		}
	}
}

func TestASagTripsOnTheEventClockAndAShortOneRidesThrough(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		fault   string
		tripped float64
	}{{"long_sag", 1}, {"short_sag", 0}} {
		w := constWorld(t, sagSpec(t, false), model.DefaultStartTimeNS)
		if _, err := w.InjectFault("e-1", tc.fault, at(600.4), nil); err != nil {
			t.Fatal(err)
		}
		if got := w.StateValue("e-1", "voltage", at(601)); got != 100 {
			t.Fatalf("%s: voltage at the next boundary %v; a sampled trip would never see the sag", tc.fault, got)
		}
		if got := w.StateValue("e-1", "sag_trip", at(601)); got != tc.tripped {
			t.Fatalf("%s: trip %v at 601 s, want %v", tc.fault, got, tc.tripped)
		}
	}
}

func emittingWorld(t *testing.T, mutate func(*model.DomainSpec)) (*World, *[]model.SimEvent) {
	t.Helper()
	w, err := New(testSpec(t, mutate), 42, "w-test", model.DefaultStartTimeNS, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(e model.SimEvent) { events = append(events, e) })
	return w, &events
}

func TestAnEventDrivenChannelRecordsASagAndOtherStreamsDrawAsBefore(t *testing.T) {
	t.Parallel()
	sagged, sagEvents := emittingWorld(t, sagSpec(t, true))
	quiet, quietEvents := emittingWorld(t, sagSpec(t, true))
	if _, err := sagged.InjectFault("e-1", "long_sag", at(600.4), nil); err != nil {
		t.Fatal(err)
	}
	for _, w := range []*World{sagged, quiet} {
		if _, _, err := w.Advance(at(1200)); err != nil {
			t.Fatal(err)
		}
	}
	var pq []string
	for _, e := range *sagEvents {
		if e.Channel == "pq" {
			pq = append(pq, e.EventTime)
		}
	}
	if len(pq) < 3 || pq[1] != model.FormatTime(at(600.4)) || pq[2] != model.FormatTime(at(600.7)) {
		t.Fatalf("pq records at %v, want the start and end of the sag", pq)
	}
	if periodic(*sagEvents) != periodic(*quietEvents) {
		t.Fatal("the sag changed the periodic channel's draws")
	}
	for _, e := range *sagEvents {
		if e.Channel == "noisy" && (e.EventTime == model.FormatTime(at(600.4)) || e.EventTime == model.FormatTime(at(600.7))) {
			t.Fatalf("a channel that does not observe the sag recorded at its instant %s", e.EventTime)
		}
	}
}

func periodic(events []model.SimEvent) string {
	var out string
	for _, e := range events {
		if e.Channel == "ch" {
			out += fmt.Sprintf("%s=%v;", e.EventTime, e.Value)
		}
	}
	return out
}

func TestClearingAPulseOnlyShortensItAndRecordsTheNewEnd(t *testing.T) {
	t.Parallel()
	expired, _ := emittingWorld(t, sagSpec(t, true))
	id, err := expired.InjectFault("e-1", "long_sag", at(600.4), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := expired.Advance(at(900)); err != nil {
		t.Fatal(err)
	}
	if err := expired.ClearFault(id, 0); err != nil {
		t.Fatal(err)
	}
	if v := expired.StateValue("e-1", "voltage", at(650)); v != 100 {
		t.Fatalf("a clear after the pulse ended rewrote history: voltage %v at 650 s", v)
	}
	early, events := emittingWorld(t, sagSpec(t, true))
	id, err = early.InjectFault("e-1", "long_sag", at(600.4), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := early.ClearFault(id, at(600.5)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := early.Advance(at(700)); err != nil {
		t.Fatal(err)
	}
	if !recordedAt(*events, "pq", at(600.5)) {
		t.Fatal("the early clear was not recorded at its time")
	}
}

func TestAHoldThatEndsBeforeItsDurationDoesNotTrip(t *testing.T) {
	t.Parallel()
	w := protectionWorld(t, true)
	invoke(t, w, "cut", 100.5)
	invoke(t, w, "restore", 110.3)
	if got := w.StateValue("e-1", "low_flow", at(115)); got != 0 {
		t.Fatalf("a 9.8 s low flow tripped a 10 s trip: %v", got)
	}
}

func recordedAt(events []model.SimEvent, channel string, t int64) bool {
	for _, e := range events {
		if e.Channel == channel && e.EventTime == model.FormatTime(t) {
			return true
		}
	}
	return false
}
