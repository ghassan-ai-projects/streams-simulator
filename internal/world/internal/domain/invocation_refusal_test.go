package domain

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// An invocation refused at admission is marked ErrInvalidInvocation, keeps
// its own message, and leaves no trace in the world.
func TestAnInvocationRefusedAtAdmissionLeavesTheWorldAsItWas(t *testing.T) {
	t.Parallel()
	w := newTestWorld(t, testSpec(t, nil), 23, model.DefaultStartTimeNS)
	cases := map[string][4]string{
		"unknown entity":      {"act", "ghost", "c-1", `world: unknown entity "ghost"`},
		"undeclared effector": {"nope", "e-1", "c-2", `world: unknown effector "nope"`},
		"missing command id":  {"act", "e-1", "", "world: missing command_id"},
	}
	for name, tc := range cases {
		_, err := w.InvokeEffector(tc[0], tc[1], tc[2], nil, w.Clock())
		if !errors.Is(err, ErrInvalidInvocation) || err.Error() != tc[3] {
			t.Errorf("%s: err = %v, want ErrInvalidInvocation with message %q", name, err, tc[3])
		}
	}
	if calls := w.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("refused invocations were recorded: %+v", calls)
	}
}

func TestAWorldStartsWithinTheRepresentableRange(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, nil)
	for _, start := range []int64{0, MaxStartNS} {
		if _, err := New(spec, 1, "", start, Options{}); err != nil {
			t.Fatalf("start %d refused: %v", start, err)
		}
	}
	for _, start := range []int64{-1, MaxStartNS + 1, math.MaxInt64} {
		if _, err := New(spec, 1, "", start, Options{}); err == nil || !strings.Contains(err.Error(), "is outside [0,") {
			t.Fatalf("start %d: err = %v, want refused", start, err)
		}
	}
}
