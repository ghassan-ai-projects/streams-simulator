package world

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
