package world

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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

func TestObservedTimesAreStrictlyIncreasingAcrossEmissionTies(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Entities.Count.Default = 2
		s.Faults[0].Observability.Detector.Channel = "slow"
		s.Channels = []model.Channel{
			{
				Name: "slow", ValueType: "number", Unit: "u", Resolution: 0.01,
				Fidelity: "F0", Absence: "signal", Observes: "x",
				Cadence:   model.Cadence{Mode: "periodic", PeriodS: 60},
				Noise:     model.Noise{Model: "none"},
				LinkDelay: &model.LinkDelay{Model: "constant", MeanS: 10},
			},
			{
				Name: "fast", ValueType: "number", Unit: "u", Resolution: 0.01,
				Fidelity: "F0", Absence: "signal", Observes: "x",
				Cadence:   model.Cadence{Mode: "periodic", PeriodS: 60},
				Noise:     model.Noise{Model: "none"},
				LinkDelay: &model.LinkDelay{Model: "constant", MeanS: 1},
			},
		}
	})
	start := model.DefaultStartTimeNS
	w, err := New(spec, 1, "w-observed-order", start, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(ev model.SimEvent) { events = append(events, ev) })
	if _, _, err := w.Advance(start + 60*secondsPerNS); err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 {
		t.Fatalf("expected four simultaneous-channel events, got %d", len(events))
	}
	previous := int64(0)
	for i, ev := range events {
		observed, err := model.ParseTime(ev.ObservedTime)
		if err != nil {
			t.Fatal(err)
		}
		if i > 0 && observed <= previous {
			t.Fatalf("event %d observed_time %s is not strictly after %s", i, ev.ObservedTime, model.FormatTime(previous))
		}
		previous = observed
	}
	if got := previous - (start + 60*secondsPerNS + 10*secondsPerNS); got != 3 {
		t.Fatalf("expected three one-nanosecond serialization steps after the slow candidate, got %d ns", got)
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
