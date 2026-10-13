package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func observationFault(id, form string, value, factor float64, quality string) model.Fault {
	o := &model.ObservationFault{Channel: "ch", Form: form, Quality: quality}
	if form == model.ObservationStuck {
		o.Value = &value
	}
	if form == model.ObservationScale || form == model.ObservationNoise {
		o.Factor = &factor
	}
	return model.Fault{ID: id, Onset: model.FaultOnset{Shape: "step", Magnitude: 1}, Observation: o,
		Observability: model.Observability{Detector: model.Detector{Form: model.DetectorSingleChannel, Channel: "ch"}}}
}

func instrumentSpec(t *testing.T) func(*model.DomainSpec) {
	t.Helper()
	return func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State[1].Initial = 1
		kind.Faults = append(kind.Faults,
			observationFault("stuck_high", model.ObservationStuck, 9.876, 0, "out_of_range"),
			observationFault("frozen", model.ObservationFreeze, 0, 0, ""),
			observationFault("dropped", model.ObservationDropout, 0, 0, ""),
			observationFault("gain_error", model.ObservationScale, 0, 1.5, "suspect"),
			observationFault("noisy", model.ObservationNoise, 0, 4, ""),
		)
	}
}

func instrumentWorld(t *testing.T, faultID string, onsetS float64) *World {
	t.Helper()
	w := constWorld(t, instrumentSpec(t), model.DefaultStartTimeNS)
	if faultID != "" {
		if _, err := w.InjectFault("e-1", faultID, atSeconds(onsetS), nil); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

func TestEachObservationFaultChangesOnlyWhatTheInstrumentReports(t *testing.T) {
	t.Parallel()
	clean := instrumentWorld(t, "", 0)
	at, onset := atSeconds(1800), 600.0
	cases := []struct {
		fault string
		want  float64
	}{
		{"stuck_high", 9.88},
		{"frozen", clean.Reading("e-1", "ch", atSeconds(onset))},
		{"gain_error", quantize(1.5*clean.StateValue("e-1", "x", at), 0.01)},
		{"noisy", clean.Reading("e-1", "ch", at)},
		{"dropped", clean.Reading("e-1", "ch", at)},
	}
	for _, tc := range cases {
		w := instrumentWorld(t, tc.fault, onset)
		if before := w.Reading("e-1", "ch", atSeconds(300)); before != instrumentWorld(t, "", 0).Reading("e-1", "ch", atSeconds(300)) {
			t.Fatalf("%s acts before its onset", tc.fault)
		}
		if got := w.Reading("e-1", "ch", at); math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("%s: reading %v, want %v", tc.fault, got, tc.want)
		}
		if state := w.StateValue("e-1", "x", at); state != clean.StateValue("e-1", "x", at) {
			t.Fatalf("%s changed the process: %v vs %v", tc.fault, state, clean.StateValue("e-1", "x", at))
		}
	}
	if noisy := instrumentWorld(t, "noisy", onset); noisy.ReadingNoise("e-1", "ch", at) != 0.4 || clean.ReadingNoise("e-1", "ch", at) != 0.1 {
		t.Fatalf("noise fault sigma %v, clean %v", noisy.ReadingNoise("e-1", "ch", at), clean.ReadingNoise("e-1", "ch", at))
	}
}

func emittedEvents(t *testing.T, faultID string, onsetS, untilS float64) []model.SimEvent {
	t.Helper()
	spec := testSpec(t, instrumentSpec(t))
	w, err := New(spec, 5, "w-obs", model.DefaultStartTimeNS, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.InjectFault("e-1", faultID, atSeconds(onsetS), nil); err != nil {
		t.Fatal(err)
	}
	var events []model.SimEvent
	w.SetEmitter(func(e model.SimEvent) { events = append(events, e) })
	if _, _, err := w.Advance(atSeconds(untilS)); err != nil {
		t.Fatal(err)
	}
	return events
}

func TestADroppedOutChannelEmitsNothingAndOthersReportTheirQuality(t *testing.T) {
	t.Parallel()
	if events := emittedEvents(t, "dropped", 600, 1800); len(events) != 9 {
		t.Fatalf("a channel dropping out at 600 s emits only its 9 readings before the onset, got %d", len(events))
	}
	for _, tc := range []struct{ fault, quality string }{{"stuck_high", "out_of_range"}, {"gain_error", "suspect"}, {"frozen", ""}} {
		for _, e := range emittedEvents(t, tc.fault, 600, 1800) {
			at, err := model.ParseTime(e.EventTime)
			if err != nil {
				t.Fatal(err)
			}
			if want := map[bool]string{true: tc.quality, false: ""}[at >= atSeconds(600)]; e.Quality != want {
				t.Fatalf("%s: event at %s has quality %q, want %q", tc.fault, e.EventTime, e.Quality, want)
			}
		}
	}
}

func TestAFrozenReadingIsCapturedAtOnsetWhateverTheReadSchedule(t *testing.T) {
	t.Parallel()
	events := emittedEvents(t, "frozen", 630, 3600)
	want := instrumentWorld(t, "", 0).Reading("e-1", "ch", atSeconds(630))
	for _, e := range events[10:] {
		if e.Value == nil {
			t.Fatal("a frozen channel still reports")
		}
		if e.Value != want {
			t.Fatalf("a frozen channel emitted %v at %s, want %v", e.Value, e.EventTime, want)
		}
	}
	late := instrumentWorld(t, "frozen", 630)
	if got := late.Reading("e-1", "ch", atSeconds(630)); got != want {
		t.Fatalf("a world read first at onset freezes at %v, want %v", got, want)
	}
}
