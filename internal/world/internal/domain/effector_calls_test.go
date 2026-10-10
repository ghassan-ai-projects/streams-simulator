package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func callLogSpec(t *testing.T, interlock *model.Interlock) *World {
	t.Helper()
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Effectors = []model.Effector{{
			Name:               "act",
			ArgsSchema:         map[string]any{"type": "object"},
			Ack:                model.Ack{LatencyMS: model.Latency{Mean: 100}},
			Effect:             model.Effect{StateDeltas: []model.StateDelta{{State: "x", Delta: 1}}, TimeConstantS: 100},
			IdempotencyWindowS: 60,
			Interlock:          interlock,
		}}
	})
	return newTestWorld(t, spec, 5, model.DefaultStartTimeNS)
}

// The call log is the scoring authority: each invocation yields exactly one
// record carrying the request identity and what became of it.
func TestInvocationIsRecordedWithItsIdentityAndOutcome(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	args := map[string]any{"k": "v"}
	result, err := w.InvokeEffector("act", "e-1", "cmd-1", args, at)
	if err != nil {
		t.Fatal(err)
	}
	calls := w.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	c := calls[0]
	if c.CommandID != "cmd-1" || c.Effector != "act" || c.EntityID != "e-1" || c.AtNS != at || c.WorldID != w.ID {
		t.Fatalf("identity not recorded: %+v", c)
	}
	if c.Mode != result.Mode || c.Accepted != result.Accepted || c.EffectApplied != result.EffectApplied || c.AckLatencyMS != result.AckLatencyMS {
		t.Fatalf("outcome %+v does not match result %+v", c, result)
	}
	if c.ResultDigest == "" || c.InterlockRefused {
		t.Fatalf("digest/interlock wrong: %+v", c)
	}
}

func TestInterlockRefusalIsRecordedAsRejectedWithItsReason(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, &model.Interlock{State: "u", Operator: "gt", Threshold: 1})
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at); !errors.Is(err, ErrInterlockRefused) {
		t.Fatalf("err = %v, want ErrInterlockRefused", err)
	}
	calls := w.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	c := calls[0]
	if !c.InterlockRefused || c.Accepted || c.EffectApplied || c.Mode != ModeReject || c.Reason != "interlock_refused" {
		t.Fatalf("refusal record = %+v", c)
	}
}

func TestIdempotentReplayAddsNoSecondCallRecord(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	first, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at)
	if err != nil {
		t.Fatal(err)
	}
	again, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at+secondsPerNS)
	if err != nil {
		t.Fatal(err)
	}
	// The replay reconstructs the result from the call record, which does not
	// keep the effect ETA (DEFERRED D-42), so only the recorded fields compare.
	same := first.Accepted == again.Accepted && first.CommandID == again.CommandID &&
		first.Mode == again.Mode && first.EffectApplied == again.EffectApplied
	if !same || len(w.EffectorCalls()) != 1 {
		t.Fatalf("replay differs or logged twice: %+v vs %+v, %d calls", first, again, len(w.EffectorCalls()))
	}
	later, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at+61*secondsPerNS)
	if err != nil || len(w.EffectorCalls()) != 2 || later.CommandID != "cmd-1" {
		t.Fatalf("expired window must execute again: %v, %d calls", err, len(w.EffectorCalls()))
	}
}

// A refusal is terminal but not cached: repeating the same command id is
// evaluated, and refused, again.
func TestInterlockRefusalIsNotCachedForIdempotentReplay(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, &model.Interlock{State: "u", Operator: "gt", Threshold: 1})
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	for i := 0; i < 2; i++ {
		if _, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{}, at); !errors.Is(err, ErrInterlockRefused) {
			t.Fatalf("attempt %d: err = %v, want ErrInterlockRefused", i, err)
		}
	}
	if n := len(w.EffectorCalls()); n != 2 {
		t.Fatalf("calls = %d, want one record per refusal", n)
	}
}

// A replayed command answers with exactly what it first answered, including
// when the effect becomes visible.
func TestIdempotentReplayReturnsTheFirstResultIncludingItsEffectETA(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	first, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": "v"}, at)
	if err != nil {
		t.Fatal(err)
	}
	again, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": "v"}, at+secondsPerNS)
	if err != nil {
		t.Fatal(err)
	}
	if *again != *first {
		t.Fatalf("replay differs from the first answer:\nfirst: %+v\nagain: %+v", *first, *again)
	}
	if first.EffectETANS == 0 {
		t.Fatalf("the first answer must carry an effect ETA: %+v", first)
	}
	if got := len(w.EffectorCalls()); got != 1 {
		t.Fatalf("a replay must not add a call record, got %d", got)
	}
}

// The same command_id for another request must not be acknowledged with the
// first request's answer: nothing the caller asked for happened.
func TestCommandIDReusedForADifferentRequestIsRefused(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": "v"}, at); err != nil {
		t.Fatal(err)
	}
	for name, call := range map[string]func() error{
		"other arguments": func() error {
			_, err := w.InvokeEffector("act", "e-1", "cmd-1", map[string]any{"k": "other"}, at)
			return err
		},
	} {
		if err := call(); !errors.Is(err, ErrCommandIDReused) {
			t.Errorf("%s: err = %v, want ErrCommandIDReused", name, err)
		}
	}
	if got := len(w.EffectorCalls()); got != 1 {
		t.Fatalf("a refused reuse must not add a call record, got %d", got)
	}
}

// The call log is the scoring authority: nobody who issued a command, or who
// reads the log, can rewrite what it says was asked.
func TestEffectorCallLogIsIsolatedFromCallersArguments(t *testing.T) {
	t.Parallel()
	w := callLogSpec(t, nil)
	at := model.DefaultStartTimeNS + 5*secondsPerNS
	args := map[string]any{"k": "v", "nested": map[string]any{"n": 1.0}, "list": []any{"a"}}
	if _, err := w.InvokeEffector("act", "e-1", "cmd-1", args, at); err != nil {
		t.Fatal(err)
	}
	args["k"] = "changed"
	args["nested"].(map[string]any)["n"] = 2.0
	args["list"].([]any)[0] = "z"

	read := w.EffectorCalls()
	read[0].Args["k"] = "scribbled"
	again := w.EffectorCalls()[0].Args
	nested, _ := again["nested"].(map[string]any)
	list, _ := again["list"].([]any)
	if again["k"] != "v" || nested["n"] != 1.0 || list[0] != "a" {
		t.Fatalf("the logged arguments changed: %v", again)
	}
}
