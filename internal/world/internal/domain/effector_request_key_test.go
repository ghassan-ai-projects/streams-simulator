package domain

import (
	"errors"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Omitted arguments and an empty object ask for the same thing, so a retry
// that adds or drops the empty object is a replay, not a different request.
func TestOmittedArgumentsAndAnEmptyObjectAreOneRequest(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	first, err := w.InvokeEffector("act", "e-1", "cmd-1", nil, at)
	if err != nil {
		t.Fatal(err)
	}
	again, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at)
	if err != nil {
		t.Fatalf("a retry with an empty argument object must replay: %v", err)
	}
	if *again != *first {
		t.Fatalf("replay differs from the first answer:\nfirst: %+v\nagain: %+v", *first, *again)
	}
}

// Arguments json cannot encode still identify their request: two different
// requests never share a key.
func TestUnencodableArgumentsStillDistinguishRequests(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": math.NaN()}, at); err != nil {
		t.Fatal(err)
	}
	_, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": math.NaN(), "other": 1.0}, at)
	if !errors.Is(err, ErrCommandIDReused) {
		t.Fatalf("err = %v, want ErrCommandIDReused", err)
	}
}
