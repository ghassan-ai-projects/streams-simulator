package domain

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAnRelayWithoutInterruptsShowsAnEmptyView(t *testing.T) {
	t.Parallel()
	if view := (&Relay{}).InterruptView(); !reflect.DeepEqual(view, EmptyInterruptView()) {
		t.Fatalf("view = %v, want the empty view", view)
	}
}

func TestTheWorldClockOfAnInterruptedRelayOnlyMovesForward(t *testing.T) {
	t.Parallel()
	r := &Relay{}
	r.EnableInterrupts()
	r.noteWorldTime(Inbound{Message: Message{WorldTimeNS: 7}})
	r.noteWorldTime(Inbound{Message: Message{WorldTimeNS: 3}})
	if r.ints.simNS != 7 {
		t.Fatalf("world clock = %d, want 7 kept over the earlier 3", r.ints.simNS)
	}
	plain := &Relay{}
	plain.noteWorldTime(Inbound{Message: Message{WorldTimeNS: 9}})
	if plain.ints != nil {
		t.Fatal("a relay without interrupts must not grow an interrupt state")
	}
}

func TestAnAnswerIsNamedByItsReasonOrByTheReportedOutcomeOfAResult(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   Inbound
		want string
	}{
		{"a receipt with a reason", Inbound{Message: Message{Type: "DRECEIPT", Status: "refused", Reason: "busy"}}, "refused:busy"},
		{"a receipt with no reason", Inbound{Message: Message{Type: "DRECEIPT", Status: "accepted"}}, "accepted"},
		{"a result", Inbound{Message: Message{Type: "DRESULT", Status: "ok", ReportedOutcome: "succeeded", Reason: "ignored"}}, "succeeded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := answerText(tc.in); got != tc.want {
				t.Fatalf("answerText = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestADeviceWithNoBootIsQueuedAsNoneAndASafeStopNeedsADeclaredOne(t *testing.T) {
	t.Parallel()
	if got := (&Relay{}).queuedBoot("ghost"); got != noBoot {
		t.Fatalf("queuedBoot(unborn) = %q, want %q", got, noBoot)
	}
	kinds := []any{map[string]any{"kind": "fan", "commands": []any{map[string]any{"name": "run", "safe_stop": false}}}}
	if got := safeStopOf(kinds, "fan"); got != "" {
		t.Fatalf("safeStopOf(no declared safe stop) = %q, want empty", got)
	}
}

func TestAStopForADeviceThatWasNeverBornIsRecordedAsNotBorn(t *testing.T) {
	t.Parallel()
	r := &Relay{}
	r.EnableInterrupts()
	var moves Moves
	r.issueStop(&moves, StopOrder{Device: "ghost", Command: "halt"})
	if len(moves.Interrupt) != 1 || moves.Interrupt[0].Event() != EventEffect ||
		moves.Interrupt[0]["outcome"] != stopNotBorn || moves.Interrupt[0]["device"] != "ghost" {
		t.Fatalf("interrupts = %v, want one effect that the device was not born", moves.Interrupt)
	}
}

func TestAStopForASimulatedDeviceAfterTheSimulatorIsLostIsRecordedAsLost(t *testing.T) {
	t.Parallel()
	r := &Relay{sim: &simSide{devices: map[string]bool{"S1": true}, born: map[string]string{"S1": "sim-boot"}, lost: true}}
	r.EnableInterrupts()
	var moves Moves
	r.issueStop(&moves, StopOrder{Device: "S1", Command: "halt"})
	if len(moves.Interrupt) != 1 || moves.Interrupt[0]["outcome"] != stopSimulatorGone || len(moves.ToSim) != 0 {
		t.Fatalf("moves = %+v, want one effect that the simulator is gone and nothing sent to it", moves)
	}
}

func TestTheIssuedStopTableStartsAgainWhenItIsFull(t *testing.T) {
	t.Parallel()
	r := &Relay{}
	r.EnableInterrupts()
	for i := range maxIssuedStops {
		r.ints.issued[fmt.Sprintf("filled-%d", i)] = "d"
	}
	before := len(r.ints.issued)
	record := r.stopRecord(StopOrder{Device: "P1", Command: "halt"}, OriginGateway, "boot-1")
	if before != maxIssuedStops || len(r.ints.issued) != 1 || r.ints.orders != 1 {
		t.Fatalf("issued = %d (before %d) orders = %d, want the full table cleared to the one new stop", len(r.ints.issued), before, r.ints.orders)
	}
	if record.Fields["command_id"] != stopIDPrefix+"1" || r.ints.issued[stopIDPrefix+"1"] != "P1" {
		t.Fatalf("record = %v, want the first stop id issued for P1", record.Fields["command_id"])
	}
}

func TestAHeldCommandAfterApprovalOfASimulatedDeviceGoesDownToTheWorld(t *testing.T) {
	t.Parallel()
	g := interruptMixedRig(t)
	control(t, g.relay, approveModeLine(ScopeDevice, "S1", true))
	queued := g.hostSends(commandFor("S1", "c-9"))
	receiptOf(t, queued.Up, "queued", "")
	reply, approved := control(t, g.relay, `{"op":"approve","command_id":"c-9"}`)
	if reply["result"] != ResultApplied || len(approved.ToSim) != 1 || approved.ToSim[0]["device"] != "S1" {
		t.Fatalf("reply %v to the world %v, want the approved command for S1 sent to the simulator", reply, approved.ToSim)
	}
}
