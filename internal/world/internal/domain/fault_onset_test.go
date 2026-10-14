package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAFaultOnsetBeforeTheClockIsRefusedAndNothingIsRegistered(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 20e9); err != nil {
		t.Fatal(err)
	}
	_, err := w.InjectFault("e-1", "f1", model.DefaultStartTimeNS+10e9, nil)
	if !errors.Is(err, ErrFaultOnsetInPast) {
		t.Fatalf("an onset before the clock: err = %v, want ErrFaultOnsetInPast", err)
	}
	if faults := w.ListFaults(); len(faults) != 0 {
		t.Fatalf("a refused injection registered %v", faults)
	}
}

func TestAFaultOnsetAtOrAfterTheClockIsAccepted(t *testing.T) {
	t.Parallel()
	clock := model.DefaultStartTimeNS + 20e9
	for name, onset := range map[string]int64{"now": clock, "unset": 0, "future": clock + 5e9} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			w := constWorld(t, nil, model.DefaultStartTimeNS)
			if _, _, err := w.Advance(clock); err != nil {
				t.Fatal(err)
			}
			if _, err := w.InjectFault("e-1", "f1", onset, nil); err != nil {
				t.Fatalf("onset %d at clock %d: %v", onset, clock, err)
			}
		})
	}
}

func TestAFaultAlreadyUnderwayAtTheStartOfTheRunIsAccepted(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	if _, err := w.InjectFault("e-1", "f1", model.DefaultStartTimeNS-600e9, nil); err != nil {
		t.Fatalf("a pre-degraded onset before the start of an unrun world: %v", err)
	}
}

func TestAFaultClearBeforeTheClockIsRefusedAndTheFaultStaysActive(t *testing.T) {
	t.Parallel()
	w := constWorld(t, nil, model.DefaultStartTimeNS)
	id, err := w.InjectFault("e-1", "f1", model.DefaultStartTimeNS+5e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(model.DefaultStartTimeNS + 20e9); err != nil {
		t.Fatal(err)
	}
	if err := w.ClearFault(id, model.DefaultStartTimeNS+10e9); !errors.Is(err, ErrFaultClearInPast) {
		t.Fatalf("a clear before the clock: err = %v, want ErrFaultClearInPast", err)
	}
	if faults := w.ListFaults(); len(faults) != 1 {
		t.Fatalf("faults after a refused clear = %v, want the fault still active", faults)
	}
	if err := w.ClearFault(id, 0); err != nil || len(w.ListFaults()) != 0 {
		t.Fatalf("a clear now: %v, faults %v", err, w.ListFaults())
	}
}
