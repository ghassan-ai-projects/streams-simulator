package domain

import (
	"fmt"
	"testing"
)

func TestACommandInApprovalModeWaitsAndTheOperatorApprovesIt(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeDevice, "pump-1", true))
	queued := judge(g, "c-1", "pump-1")
	if len(queued.Queued) != 1 || queued.Queued[0].CommandID != "c-1" || queued.Proceed {
		t.Fatalf("outcome = %+v, want the command queued and not passed", queued)
	}
	if pending := g.View()["pending"].([]any); len(pending) != 1 || pending[0].(map[string]any)["command_id"] != "c-1" {
		t.Fatalf("pending = %v, want c-1 waiting", pending)
	}
	approved := apply(t, g, `{"op":"approve","command_id":"c-1"}`)
	if len(approved.Forward) != 1 || approved.Forward[0].CommandID != "c-1" || approved.Forward[0].IdempotencyKey != dcmdOf("c-1", "pump-1").IdempotencyKey {
		t.Fatalf("forward = %+v, want the command exactly as it arrived", approved.Forward)
	}
	if approved.Reply["result"] != ResultApplied || len(g.View()["pending"].([]any)) != 0 {
		t.Fatalf("reply %v pending %v, want it applied and nothing left waiting", approved.Reply, g.View()["pending"])
	}
}

func TestADeniedCommandIsRefusedAsLocalControlAndAWaitingOneExpiresAtItsDeadline(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-deny", "pump-1")
	judge(g, "c-late", "pump-2")
	denied := apply(t, g, `{"op":"deny","command_id":"c-deny"}`)
	if len(denied.Refused) != 1 || denied.Refused[0].CommandID != "c-deny" || denied.Refused[0].Reason != ReasonLocalControl || len(denied.Forward) != 0 {
		t.Fatalf("denied = %+v, want c-deny refused as local_control", denied)
	}
	if early := g.Advance(ClockWorld, deadlineNS); len(early.Refused) != 0 {
		t.Fatalf("at the deadline = %+v, want the command to wait: its deadline has not passed", early)
	}
	if other := g.Advance(ClockWall, deadlineNS+1); len(other.Refused) != 0 {
		t.Fatalf("on another clock = %+v, want a world command to expire only on the world clock", other)
	}
	late := g.Advance(ClockWorld, deadlineNS+1)
	if len(late.Refused) != 1 || late.Refused[0].Reason != "expired" || late.Refused[0].CommandID != "c-late" || len(g.View()["pending"].([]any)) != 0 {
		t.Fatalf("late = %+v, want c-late refused as expired and nothing left waiting", late)
	}
}

func TestACommandAlreadyPastItsDeadlineIsRefusedAsExpiredWithoutWaiting(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	outcome := g.Judge(dcmdOf("c-1", "pump-1"), false, ClockWorld, deadlineNS+1)
	if got := verdictFor(outcome); got != "refuse:expired" || len(g.View()["pending"].([]any)) != 0 {
		t.Fatalf("verdict %s pending %v, want expired and nothing waiting", got, g.View()["pending"])
	}
}

func TestAHoldSetWhileACommandWaitsRefusesItAtOnce(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-1", "pump-1")
	judge(g, "c-2", "valve-1")
	held := apply(t, g, scopeLine(OpHold, ScopeBoard, "b-1"))
	if len(held.Refused) != 1 || held.Refused[0].CommandID != "c-1" || held.Refused[0].Reason != ReasonLocalControl {
		t.Fatalf("held = %+v, want c-1 refused and c-2, on another board, left waiting", held)
	}
	pending := g.View()["pending"].([]any)
	if len(pending) != 1 || pending[0].(map[string]any)["command_id"] != "c-2" {
		t.Fatalf("pending = %v, want c-2 alone", pending)
	}
}

func TestAnApprovalOfACommandThatAHoldNowCoversIsRefused(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-1", "valve-1")
	g.machine.holds[Scope{Kind: ScopeDevice, ID: "valve-1"}] = false
	approved := apply(t, g, `{"op":"approve","command_id":"c-1"}`)
	if len(approved.Forward) != 0 || len(approved.Refused) != 1 || approved.Refused[0].Reason != ReasonLocalControl {
		t.Fatalf("approved = %+v, want the approval to be refused as local_control: a hold outranks it", approved)
	}
}

func TestARepeatOfAWaitingCommandWaitsToAndOtherContentUnderItsKeyIsADuplicate(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-1", "pump-1")
	again := dcmdOf("c-2", "pump-1")
	if got := verdictFor(g.Judge(again, false, ClockWorld, nowNS)); got != VerdictQueue {
		t.Fatalf("a repeat = %s, want it queued", got)
	}
	other := dcmdOf("c-3", "pump-2")
	other.IdempotencyKey = again.IdempotencyKey
	if got := verdictFor(g.Judge(other, false, ClockWorld, nowNS)); got != "refuse:duplicate" {
		t.Fatalf("other content under the key = %s, want a duplicate", got)
	}
	approved := apply(t, g, `{"op":"approve","command_id":"c-1"}`)
	if len(approved.Forward) != 2 || approved.Forward[0].CommandID != "c-1" || approved.Forward[1].CommandID != "c-2" {
		t.Fatalf("forward = %+v, want the command and its repeat in the order they arrived", approved.Forward)
	}
}

func TestThePendingQueueAndTheRepeatsOfOneCommandAreBounded(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	for i := range maxPendingApprovals {
		message := dcmdOf(fmt.Sprintf("c-%03d", i), "pump-1")
		message.Args = map[string]any{"speed": i}
		message.IdempotencyKey = CommandOf(message.Message).Content()
		if got := verdictFor(g.Judge(message, false, ClockWorld, nowNS)); got != VerdictQueue {
			t.Fatalf("command %d = %s, want it queued", i, got)
		}
	}
	full := dcmdOf("c-full", "pump-1")
	full.Args = map[string]any{"speed": 99}
	full.IdempotencyKey = CommandOf(full.Message).Content()
	if got := verdictFor(g.Judge(full, false, ClockWorld, nowNS)); got != "refuse:not_ready" {
		t.Fatalf("the command over the bound = %s, want not_ready", got)
	}
	first := dcmdOf("c-000", "pump-1")
	first.Args = map[string]any{"speed": 0}
	first.IdempotencyKey = CommandOf(first.Message).Content()
	for i := range maxWaitingRepeats {
		first.CommandID = fmt.Sprintf("r-%d", i)
		if got := verdictFor(g.Judge(first, false, ClockWorld, nowNS)); got != VerdictQueue {
			t.Fatalf("repeat %d = %s, want it queued", i, got)
		}
	}
	first.CommandID = "r-over"
	if got := verdictFor(g.Judge(first, false, ClockWorld, nowNS)); got != "refuse:not_ready" {
		t.Fatalf("the repeat over the bound = %s, want not_ready", got)
	}
}

func TestEndingTheSessionRefusesWhatWaitsAsSessionEndingAndLogsNothingWhenNothingWaits(t *testing.T) {
	t.Parallel()
	g := plantGate()
	if empty := g.Flush(ClockWorld, nowNS); len(empty.Entries) != 0 {
		t.Fatalf("flush of nothing logged %v", empty.Entries)
	}
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-1", "pump-1")
	flushed := g.Flush(ClockWorld, nowNS)
	if len(flushed.Refused) != 1 || flushed.Refused[0].Reason != "session_ending" || len(g.View()["pending"].([]any)) != 0 {
		t.Fatalf("flushed = %+v, want c-1 refused as session_ending", flushed)
	}
}
