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
	reply, moves := estopOf(t, g, "P1")
	if unreachable := reply["unreachable"].([]any); reply["result"] != ResultApplied || !slices.Equal(unreachable, []any{"P1"}) || len(moves.Down) != 0 {
		t.Fatalf("reply %v down %v, want a device of a gateway that is gone reported unreachable, never vanished or refused as unknown", reply, typesOf(moves.Down))
	}
	control(t, g.relay, scopeLine(OpRelease, ScopeDevice, "P1"))
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(stoppableBirth(t))
	reply, moves = estopOf(t, g, "P1")
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

func TestAnEmergencyStopNamesADeviceThatWentOfflineAsUnreachableForEveryScopeThatCoversItAndStopsTheRest(t *testing.T) {
	t.Parallel()
	for name, line := range map[string]string{"all": scopeLine(OpEStop, ScopeAll, ""), "its board": scopeLine(OpEStop, ScopeBoard, "b-1"),
		"itself": scopeLine(OpEStop, ScopeDevice, "P1")} {
		g := interruptMixedRig(t)
		g.relay.GatewayLost()
		reply, moves := control(t, g.relay, line)
		unreachable, stopped := reply["unreachable"].([]any), reply["stopped"].([]any)
		if !slices.Equal(unreachable, []any{"P1"}) {
			t.Errorf("%s: reply %v, want the offline gateway device P1 in unreachable and in no other list", name, reply)
		}
		if name == "all" && (!slices.Equal(stopped, []any{"S1"}) || len(moves.Down)+len(moves.ToSim) == 0) {
			t.Errorf("%s: reply %v, want the simulated device still stopped", name, reply)
		}
	}
}

func TestAHoldScopedToADeviceThatWentOfflineIsAcceptedAndListedInTheState(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.relay.GatewayLost()
	reply, _ := control(t, g.relay, scopeLine(OpHold, ScopeDevice, "P1"))
	holds := g.relay.InterruptView()["holds"].([]any)
	if reply["result"] != ResultApplied || len(holds) != 1 || holds[0].(map[string]any)["id"] != "P1" {
		t.Fatalf("reply %v holds %v, want a hold on an announced device accepted although it is offline", reply, holds)
	}
	if reply, _ := control(t, g.relay, scopeLine(OpHold, ScopeDevice, "ghost")); reply["reason"] != ReasonUnknownDevice {
		t.Fatalf("reply = %v, want a device nobody announced still refused as unknown", reply)
	}
}

func TestADeviceAnEarlierBirthAnnouncedStaysInTheInventoryAfterALaterBirthLeavesItOut(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	control(t, g.relay, scopeLine(OpHold, ScopeDevice, "V1"))
	g.relay.GatewayLost()
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(deviceOnlyBirth(t))
	reply, _ := estopOf(t, g, "V1")
	if reply["result"] != ResultApplied {
		t.Fatalf("reply = %v, want V1, announced earlier in this session, still known after a birth without it", reply)
	}
}

func estopEpochAfter(t *testing.T, g *relayRig) any {
	t.Helper()
	_, moves := estopOf(t, g, "P1")
	if len(moves.Down) != 1 {
		t.Fatalf("down %v, want one stop for P1", typesOf(moves.Down))
	}
	return moves.Down[0]["authority_epoch"]
}

func TestAHostThatReturnsWithoutAnEpochNeverMakesTheStopCarryEpochZero(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.relay.HostLost()
	g.relay.HostAttached()
	g.newHostSession("relay-h2")
	g.hostSends(hostState(0, 0))
	if epoch := estopEpochAfter(t, g); epoch != int64(7) {
		t.Fatalf("authority_epoch = %v, want the last epoch a host announced (7) while the new host has announced none", epoch)
	}
}

func TestALostHostThatNeverAnnouncedAnEpochDoesNotEraseTheEpochAnEarlierHostGave(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	g.relay.HostLost()
	g.relay.HostAttached()
	g.newHostSession("relay-h2")
	g.hostSends(hostState(0, 0))
	g.relay.HostLost()
	g.gw(stoppableBirth(t))
	if epoch := estopEpochAfter(t, g); epoch != int64(7) {
		t.Fatalf("authority_epoch = %v, want 7 kept: losing a host that announced epoch 0 must not overwrite it", epoch)
	}
}
