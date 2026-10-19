package domain

import (
	"slices"
	"testing"
)

func TestAHostCommandUsingTheReservedStopPrefixIsRefusedBeforeItCanBeConfusedWithTheNodesOwn(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	moves := g.hostSends(commandFor("P1", "node-stop-1"))
	receiptOf(t, moves.Up, "refused", ReasonMalformed)
	if len(moves.Down) != 0 || len(moves.Interrupt) != 0 {
		t.Fatalf("down %v log %v, want the reserved id refused before the rules or the gateway saw it", typesOf(moves.Down), eventsOf(moves.Interrupt))
	}
}

func TestEveryStopTheRelayOrdersHasItsOwnIdAndIdempotencyKeyAndIsRecordedAsOrdered(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	var stops []map[string]any
	for range 2 {
		_, moves := control(t, g.relay, scopeLine(OpEStop, ScopeDevice, "P1"))
		stops = append(stops, moves.Down...)
		if ordered := eventsOf(moves.Interrupt); !slices.Contains(ordered, EventEffect) {
			t.Fatalf("events = %v, want the order recorded as an effect, so a report can name the command", ordered)
		}
		control(t, g.relay, scopeLine(OpRelease, ScopeDevice, "P1"))
	}
	if len(stops) != 2 || stops[0]["command_id"] == stops[1]["command_id"] || stops[0]["idempotency_key"] == stops[1]["idempotency_key"] {
		t.Fatalf("stops = %v, want two commands with different ids and keys: a gateway that deduplicates by key would otherwise skip the second stop", stops)
	}
}

func TestTheTableOfIssuedStopsEvictsTheOldestInsteadOfBeingClearedWhole(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	ints := g.relay.ints
	for range maxIssuedStops + 5 {
		ints.issue("P1")
	}
	if len(ints.issued) != maxIssuedStops || len(ints.order) != maxIssuedStops {
		t.Fatalf("issued %d ids (%d ordered), want exactly the newest %d kept", len(ints.issued), len(ints.order), maxIssuedStops)
	}
	if _, kept := ints.issued["node-stop-1"]; kept {
		t.Fatal("the oldest id is still tracked after the table filled")
	}
	if _, kept := ints.issued[stopIDPrefix+"4101"]; !kept {
		t.Fatalf("the newest id is not tracked: %v", len(ints.issued))
	}
}

func estopOf(t *testing.T, g *relayRig, device string) (map[string]any, Moves) {
	t.Helper()
	return control(t, g.relay, scopeLine(OpEStop, ScopeDevice, device))
}

func TestAnEmergencyStopStillReachesTheGatewayAfterTheHostWasLostAndTheBirthsForgotten(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.relay.HostLost()
	g.gw(stoppableBirth(t))
	reply, moves := estopOf(t, g, "P1")
	if len(moves.Down) != 1 || moves.Down[0]["device"] != "P1" || moves.Down[0]["expected_boot_id"] != "boot-1" || moves.Down[0]["authority_epoch"] != int64(7) {
		t.Fatalf("down %v, want the stop sent to the gateway on the last boot it announced, with the host's last epoch", moves.Down)
	}
	if stopped := reply["stopped"].([]any); !slices.Equal(stopped, []any{"P1"}) || len(reply["unreachable"].([]any)) != 0 {
		t.Fatalf("reply = %v, want P1 stopped and nothing unreachable: losing the host does not unreach the gateway", reply)
	}
}

func TestAGatewayRebirthWithANewBootUpdatesTheBootTheStopIsSentTo(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.gw(deviceBirthOf("P1", "boot-2"))
	_, moves := estopOf(t, g, "P1")
	if len(moves.Down) != 1 || moves.Down[0]["expected_boot_id"] != "boot-2" {
		t.Fatalf("down %v, want the stop addressed to the boot the device was last born with", moves.Down)
	}
}

func TestADeviceThatDiedIsUnreachableInTheEmergencyStopReplyAndNothingIsSent(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.gw(deviceDeath("P1", "boot-1", 0))
	reply, moves := estopOf(t, g, "P1")
	if unreachable := reply["unreachable"].([]any); !slices.Equal(unreachable, []any{"P1"}) || len(reply["stopped"].([]any)) != 0 {
		t.Fatalf("reply = %v, want P1 unreachable and never stopped", reply)
	}
	if len(moves.Down) != 0 || slices.Contains(eventsOf(moves.Interrupt), EventStop) {
		t.Fatalf("down %v events %v, want no stop ordered for a device that died", typesOf(moves.Down), eventsOf(moves.Interrupt))
	}
}

func TestAGatewayThatIsLostAndReadmittedIsNotTrustedWithBootsFromBefore(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.relay.GatewayLost()
	if reply, _ := estopOf(t, g, "P1"); reply["reason"] != ReasonUnknownDevice {
		t.Fatalf("reply = %v, want a device of a gateway that is gone refused as unknown", reply)
	}
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(stoppableBirth(t))
	reply, moves := estopOf(t, g, "P1")
	if unreachable := reply["unreachable"].([]any); !slices.Equal(unreachable, []any{"P1"}) || len(moves.Down) != 0 {
		t.Fatalf("reply %v down %v, want P1 unreachable until the readmitted gateway announces it again", reply, typesOf(moves.Down))
	}
}

func TestAStopTheGatewayAnsweredIsNamedByTheInterruptLogForTheTwinReport(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	var log []InterruptEntry
	_, moves := control(t, g.relay, scopeLine(OpEStop, ScopeDevice, "P1"))
	log = append(log, moves.Interrupt...)
	log = append(log, g.gw(Receipt("node-stop-1", "accepted", "", "boot-1")).Interrupt...)
	records := make([]map[string]any, len(log))
	for i, entry := range log {
		records[i] = parsedRecord(t, entry)
	}
	stops := NodeStops(records)
	if len(stops) != 1 || stops[0].ID != "node-stop-1" || stops[0].Device != "P1" || stops[0].Name != "halt" || stops[0].Origin != OriginNode || len(stops[0].Args) != 0 {
		t.Fatalf("NodeStops = %+v, want the one stop with its id, device, declared command and origin node", stops)
	}
	if err := ReplayInterrupts(records); err != nil {
		t.Fatalf("ReplayInterrupts = %v, want the relay's own log to replay", err)
	}
}
