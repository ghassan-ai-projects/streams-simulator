package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// A state's value at a time must not depend on which other times were read
// first: the run world reads at every emission, the oracle's world never
// emits, and both have to integrate to the same numbers.
func TestStateValueDoesNotDependOnTheReadSchedule(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Dynamics[0].DTMs = 1000
		s.Dynamics[0].F1.TimeConstantS = 600
	})
	start := model.DefaultStartTimeNS
	final := start + 3*3600*secondsPerNS + 137_000_000 // off the 1 s grid
	value := func(probes []int64) float64 {
		w, err := New(spec, 4, "w-purity", start, Options{EmitDisabled: true})
		if err != nil {
			t.Fatal(err)
		}
		for _, at := range probes {
			_ = w.Reading("e-1", "ch", at)
		}
		return w.StateValue("e-1", "x", final)
	}
	quiet := value(nil)
	probed := value([]int64{start + 777_000_000, start + 41*secondsPerNS + 5, start + 3600*secondsPerNS + 999, final - 1})
	if quiet != probed {
		t.Fatalf("value at the same time differs with the read schedule: %v vs %v", quiet, probed)
	}
}

// The stochastic fault envelope draws randomness per whole step, so its value
// is a function of the grid and not of the reads that preceded it.
func TestStochasticFaultEnvelopeDoesNotDependOnTheReadSchedule(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Faults[0].Onset = model.FaultOnset{Shape: "stochastic", RatePerHour: 0.8}
	})
	start := model.DefaultStartTimeNS
	onset := start + 600*secondsPerNS
	final := onset + 2*3600*secondsPerNS + 31_000_000_000
	value := func(probes []int64) float64 {
		w, err := New(spec, 4, "w-stoch", start, Options{EmitDisabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.InjectFault("e-1", "f1", onset, nil); err != nil {
			t.Fatal(err)
		}
		for _, at := range probes {
			_ = w.Reading("e-1", "ch", at)
		}
		return w.StateValue("e-1", "x", final)
	}
	quiet := value(nil)
	probed := value([]int64{onset + 13*secondsPerNS + 7, onset + 1800*secondsPerNS + 3, final - 5})
	if quiet != probed {
		t.Fatalf("stochastic fault value differs with the read schedule: %v vs %v", quiet, probed)
	}
}

// The entity a caller receives is a snapshot: scribbling on it must not move
// the world's own record of when the entity was born.
func TestEntitySnapshotCannotChangeTheWorld(t *testing.T) {
	t.Parallel()
	w := newTestWorld(t, testSpec(t, nil), 4, model.DefaultStartTimeNS)
	snapshot := w.Entity("e-1")
	if snapshot == nil {
		t.Fatal("e-1 must exist")
	}
	born := snapshot.BornNS
	snapshot.BornNS, snapshot.ID = born+999, "renamed"
	again := w.Entity("e-1")
	if again.BornNS != born || again.ID != "e-1" {
		t.Fatalf("the world's entity changed through a snapshot: %+v", again)
	}
}
