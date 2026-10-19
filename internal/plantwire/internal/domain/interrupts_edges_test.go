package domain

import (
	"encoding/json"
	"math"
	"testing"
)

func TestTheControlDecisionRefusesAnOperationItDoesNotKnowAsMalformed(t *testing.T) {
	t.Parallel()
	g := NewInterruptGate()
	outcome := g.Control(newEntry(EventControl, "op", "bogus"), ClockWall, 5)
	if outcome.Reply["result"] != ResultRefused || outcome.Reply["reason"] != ReasonMalformed || outcome.Reply["op"] != "bogus" {
		t.Fatalf("reply = %v, want a refusal of the unknown operation as malformed", outcome.Reply)
	}
}

func TestAScopeThatIsNotWellFormedIsRefusedAsMalformed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		scope Scope
	}{
		{"a device with no id", Scope{Kind: ScopeDevice}},
		{"a board with no id", Scope{Kind: ScopeBoard}},
		{"an all with an id", Scope{Kind: ScopeAll, ID: "x"}},
		{"an unknown kind", Scope{Kind: "room", ID: "r-1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := NewInterruptGate()
			outcome := g.Control(newEntry(EventControl, "op", OpHold, "scope", tc.scope.Fields()), ClockWall, 5)
			if outcome.Reply["result"] != ResultRefused || outcome.Reply["reason"] != ReasonMalformed {
				t.Fatalf("reply = %v, want refusal as malformed", outcome.Reply)
			}
		})
	}
}

func TestADeviceNotInTheInventoryKeepsItsNameWithNoBoardOrClock(t *testing.T) {
	t.Parallel()
	m := NewInterrupts()
	if got := m.deviceNamed("ghost"); got != (InterruptDevice{Device: "ghost"}) {
		t.Fatalf("deviceNamed(unknown) = %+v, want the bare name", got)
	}
	m.devices = []InterruptDevice{{Device: "P1", Board: "b-1", Clock: ClockWall}}
	if got := m.deviceNamed("P1"); got.Board != "b-1" || got.Clock != ClockWall {
		t.Fatalf("deviceNamed(P1) = %+v, want its board and clock", got)
	}
}

func TestAnIntegerOfAnyExactTypeIsReadAndAnythingElseIsZero(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value any
		want  int64
	}{
		{"int64", int64(12), 12},
		{"int", 7, 7},
		{"json number", json.Number("9"), 9},
		{"a string", "12", 0},
		{"nothing", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := exactInt(tc.value); got != tc.want {
				t.Fatalf("exactInt(%v) = %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}

func TestChangesInterruptsIsTrueOnlyForEntriesThatMoveTheInterruptState(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		entries []InterruptEntry
		want    bool
	}{
		{"an inventory", []InterruptEntry{newEntry(EventInventory)}, true},
		{"a control", []InterruptEntry{newEntry(EventControl)}, true},
		{"a resolution", []InterruptEntry{newEntry(EventResolve)}, true},
		{"a queued verdict", []InterruptEntry{newEntry(EventVerdict, "verdict", VerdictQueue)}, true},
		{"a passed verdict", []InterruptEntry{newEntry(EventVerdict, "verdict", VerdictPass)}, false},
		{"a reply and a stop", []InterruptEntry{newEntry(EventReply), newEntry(EventStop)}, false},
		{"nothing", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ChangesInterrupts(tc.entries); got != tc.want {
				t.Fatalf("ChangesInterrupts = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestTheEmptyViewHasNoHoldsApprovalsPendingOrDevices(t *testing.T) {
	t.Parallel()
	view := EmptyInterruptView()
	for _, key := range []string{"holds", "approve", "pending", "devices"} {
		if list, ok := view[key].([]any); !ok || len(list) != 0 {
			t.Fatalf("view[%q] = %v, want an empty list", key, view[key])
		}
	}
}

func TestAGateVerdictWithoutAnInboundPayloadChangesNothing(t *testing.T) {
	t.Parallel()
	g := NewInterruptGate()
	var outcome GateOutcome
	g.verdict(&outcome, newEntry(EventVerdict, "command_id", "c-1", "verdict", VerdictQueue), nil)
	if outcome.Proceed || len(outcome.Queued) != 0 || len(outcome.Refused) != 0 || len(g.payloads) != 0 {
		t.Fatalf("outcome = %+v payloads = %v, want no change without a payload", outcome, g.payloads)
	}
}

func TestAResolutionOfAnUnheldCommandForwardsAndRefusesNothing(t *testing.T) {
	t.Parallel()
	g := NewInterruptGate()
	var outcome GateOutcome
	g.resolved(&outcome, newEntry(EventResolve, "command_id", "ghost", "resolution", ResolutionForward, "waiting", []any{"ghost-2"}))
	if len(outcome.Forward) != 0 || len(outcome.Refused) != 0 {
		t.Fatalf("outcome = %+v, want nothing forwarded or refused for commands the gate never held", outcome)
	}
}

func TestALogEntryThatCannotBeCanonicalisedIsShownAsUnreadable(t *testing.T) {
	t.Parallel()
	if got := canonicalText(map[string]any{"x": math.NaN()}); got != "(unreadable)" {
		t.Fatalf("canonicalText(NaN) = %q, want (unreadable)", got)
	}
	if got := canonicalText(map[string]any{"b": "two", "a": 1}); got != `{"a":1,"b":"two"}` {
		t.Fatalf("canonicalText = %q, want sorted canonical JSON", got)
	}
}

func TestAReplyThatCannotBeCanonicalisedFallsBackToTheFixedBusyRefusal(t *testing.T) {
	t.Parallel()
	if got := string(ControlReply(map[string]any{"event": "reply", "x": math.Inf(1)})); got != fallbackReply {
		t.Fatalf("ControlReply = %s, want the fixed fallback", got)
	}
}
