package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func validData() Record {
	return DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, 2.5, "", 1)})
}

func TestARelayDropsWhatItCannotTrustWithoutForwardingItOrBreakingTheSession(t *testing.T) {
	t.Parallel()
	cases := map[string]func(record map[string]any){
		"a record that fails its schema": func(r map[string]any) { delete(r, "metrics") },
		"a record of another session":    func(r map[string]any) { r["session_id"] = "intruder" },
		"a record of another version":    func(r map[string]any) { r["protocol_version"] = 2 },
		"a record with no sequence":      func(r map[string]any) { delete(r, "seq") },
		"a record over its size limit":   func(r map[string]any) { r["padding"] = strings.Repeat("x", 70<<10) },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := newRelayRig(t)
			g.bringUp()
			moves := g.gwEdited(validData(), edit)
			if moves.Up != nil || moves.Down != nil || len(moves.Trace) != 1 || g.relay.Stats().Dropped != 1 {
				t.Fatalf("moves = %+v dropped %d, want nothing relayed and the record only traced as received", moves, g.relay.Stats().Dropped)
			}
			if next := g.gw(validData()); len(next.Up) != 1 {
				t.Fatalf("the next valid record produced %d records up, want the session unharmed", len(next.Up))
			}
		})
	}
}

func TestARecordOverTheSizeLimitIsRefusedForItsSizeAndNothingElse(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.gwEdited(validData(), func(r map[string]any) { r["padding"] = strings.Repeat("x", 70<<10) })
	if invalid := relayDecode(t, g.sent).Invalid; !errors.Is(invalid, ErrTooLarge) {
		t.Fatalf("invalid = %v, want ErrTooLarge", invalid)
	}
}

func TestARelayNeverCarriesUpARecordOfAHostToNodeType(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for _, record := range []Record{hostCommand(), hostState(7, 1), {Type: "NCMD", Lane: LaneControl, Fields: map[string]any{"command": "rebirth"}},
		{Type: "DQUERY", Lane: LaneControl, Fields: map[string]any{"device": "P1"}}, {Type: "HSYNC", Lane: LaneControl, Fields: map[string]any{"through_seq": 1}}} {
		if moves := g.gw(record); moves.Up != nil {
			t.Errorf("a %s from the gateway went up as %v, want it dropped: a gateway cannot command its host", record.Type, typesOf(moves.Up))
		}
	}
	if g.relay.Stats().Dropped != 5 {
		t.Fatalf("dropped = %d, want all five", g.relay.Stats().Dropped)
	}
}

func TestARelayNeverCarriesDownARecordOfANodeToHostType(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for _, record := range []Record{validData(), deviceBirthOf("P1", "boot-1"), NodeDeath(0, "shutdown"), Receipt("c-1", "accepted", "", "boot-1")} {
		if moves := g.hostSends(record); moves.Down != nil {
			t.Errorf("a %s from the host went down as %v, want it dropped", record.Type, typesOf(moves.Down))
		}
	}
}

func TestARelayDropsARepeatedRecordAndAsksOncePerGapForARebirth(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.gw(validData())
	if again := g.relay.FromGateway(relayDecode(t, g.sent)); again.Up != nil || again.Down != nil {
		t.Fatalf("a repeated record produced %+v, want it dropped", again)
	}
	first := g.gwEdited(validData(), func(r map[string]any) { r["seq"] = 10 })
	if got := typesOf(first.Down); len(got) != 1 || got[0] != "NCMD" || first.Down[0]["command"] != "rebirth" || len(first.Up) != 1 {
		t.Fatalf("first gap: down %v up %d, want one rebirth request and the record itself passed on", got, len(first.Up))
	}
	if second := g.gwEdited(validData(), func(r map[string]any) { r["seq"] = 20 }); second.Down != nil || len(second.Up) != 1 {
		t.Fatalf("second gap: down %v up %d, want no second request while the first is unanswered", second.Down, len(second.Up))
	}
	g.gwEdited(hardwareBirth(t, nil), func(r map[string]any) { r["seq"] = 21 })
	if third := g.gwEdited(validData(), func(r map[string]any) { r["seq"] = 30 }); len(third.Down) != 1 {
		t.Fatalf("gap after the rebirth: down %v, want a new request", third.Down)
	}
}

func TestARelayIgnoresEverythingBeforeAnAdmittedBirth(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	if moves := g.gw(validData()); moves.Up != nil || moves.Down != nil || g.relay.GatewayAdmitted() {
		t.Fatalf("data before a birth produced %+v, want it dropped", moves)
	}
	broken := g.gwEdited(hardwareBirth(t, nil), func(r map[string]any) { delete(r, "kinds") })
	if broken.Down != nil || g.relay.GatewayAdmitted() || g.relay.Stats().Dropped != 2 {
		t.Fatalf("an invalid birth produced %+v admitted %v, want silence like the reference host", broken, g.relay.GatewayAdmitted())
	}
}

func TestARelayDropsWhatAHostSendsThatIsInvalidRepeatedOrNotItsOwn(t *testing.T) {
	t.Parallel()
	cases := map[string]func(record map[string]any){
		"a command with a member the contract does not know": func(r map[string]any) { r["priority"] = "high" },
		"a command of another session":                       func(r map[string]any) { r["session_id"] = "intruder" },
		"a command that fails its schema":                    func(r map[string]any) { delete(r, "expected_boot_id") },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := newRelayRig(t)
			g.bringUp()
			if moves := g.hostEdited(hostCommand(), edit); moves.Down != nil || moves.Err != nil || g.relay.Stats().Dropped != 1 {
				t.Fatalf("moves = %+v dropped %d, want nothing handed to the gateway", moves, g.relay.Stats().Dropped)
			}
			if next := g.hostSends(hostCommand()); len(next.Down) != 1 {
				t.Fatalf("the next valid command produced %d records down, want the session unharmed", len(next.Down))
			}
		})
	}
}

func TestARelayDropsARepeatedHostRecordButLosesAHostWhoseSequenceHasAGap(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.hostSends(hostCommand())
	if again := g.relay.FromHost(relayDecode(t, g.hostSent)); again.Down != nil || again.CloseHost {
		t.Fatalf("a repeated host record produced %+v, want it dropped", again)
	}
	moves := g.hostEdited(hostCommand(), func(r map[string]any) { r["seq"] = 40 })
	if !moves.CloseHost || len(moves.Down) != 1 || moves.Down[0]["online"] != false || hasType(moves.Down, "DCMD") {
		t.Fatalf("moves = %+v, want the host lost on a gap and the late command not delivered", moves)
	}
}

func TestARelayDropsAHandshakeRecordThatIsNotAStateOrNotValid(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, nil))
	g.relay.HostAttached()
	if moves := g.hostSends(hostCommand()); moves.Down != nil || moves.CloseHost {
		t.Fatalf("a command before the handshake produced %+v, want it dropped", moves)
	}
	if moves := g.hostEdited(hostState(7, 0), func(r map[string]any) { delete(r, "host_id") }); moves.Down != nil || moves.CloseHost {
		t.Fatalf("an invalid STATE produced %+v, want it dropped", moves)
	}
	if g.relay.Stats().Dropped != 2 {
		t.Fatalf("dropped = %d, want both", g.relay.Stats().Dropped)
	}
}

func TestEveryRelayedRecordIsTracedAsReceivedThenAsSentWhileADroppedOneIsOnlyReceived(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.gwEdited(validData(), func(r map[string]any) { delete(r, "metrics") })
	moves := g.gw(validData())
	if len(moves.Trace) != 2 || moves.Trace[0].Direction != TraceGateway || moves.Trace[1].Direction != TraceNode {
		t.Fatalf("trace = %+v, want the record as received then as sent", moves.Trace)
	}
	if moves.Trace[0].Record["seq"] != json.Number("4") || moves.Trace[1].Record["seq"] != uint64(3) {
		t.Fatalf("seq received %v sent %v, want the gateway's 4 and the relay's own 3", moves.Trace[0].Record["seq"], moves.Trace[1].Record["seq"])
	}
	down := g.hostSends(hostState(7, 1))
	if len(down.Trace) != 2 || down.Trace[0].Direction != TraceHost || down.Trace[1].Direction != TraceRelay {
		t.Fatalf("trace = %+v, want the STATE as received from the host then as sent to the gateway", down.Trace)
	}
}
