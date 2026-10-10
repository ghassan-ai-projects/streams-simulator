package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Reading is the noise-free observation. On a world that does emit noise it
// must not draw from the noise streams or move the drift state, or merely
// asking would change what the run goes on to emit.
func TestReadingNeverAltersWhatARunEmits(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Channels[0].Drift = &model.Drift{Model: "random_walk", RatePerHour: 0.5}
	})
	start := model.DefaultStartTimeNS
	trace := func(withReads bool) []model.SimEvent {
		w, err := New(spec, 9, "w-read", start, Options{})
		if err != nil {
			t.Fatal(err)
		}
		var events []model.SimEvent
		w.SetEmitter(func(e model.SimEvent) { events = append(events, e) })
		for step := int64(1); step <= 10; step++ {
			if withReads {
				_ = w.Reading("e-1", "ch", start+step*30*secondsPerNS)
			}
			if _, _, err := w.Advance(start + step*60*secondsPerNS); err != nil {
				t.Fatal(err)
			}
		}
		return events
	}
	plain, probed := trace(false), trace(true)
	if len(plain) == 0 || !reflect.DeepEqual(plain, probed) {
		t.Fatalf("reads changed the emitted trace: %d events vs %d", len(plain), len(probed))
	}
}

func TestReadingOnANoisyWorldIsTheNoiseFreeValue(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, nil)
	w, err := New(spec, 9, "w-read", model.DefaultStartTimeNS, Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	at := model.DefaultStartTimeNS + 3600*secondsPerNS
	first, second := w.Reading("e-1", "ch", at), w.Reading("e-1", "ch", at)
	if first != second {
		t.Fatalf("a noise-free reading is repeatable: %v then %v", first, second)
	}
	if first == 0 {
		t.Fatal("the reading must reflect the hidden state")
	}
}
