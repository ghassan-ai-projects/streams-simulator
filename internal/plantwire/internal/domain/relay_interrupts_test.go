package domain

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const hostDeadlineNS = int64(1767225660000000000)

func stoppableKinds() []any {
	command := func(name string, safeStop bool) map[string]any {
		return map[string]any{"name": name, "args_schema": map[string]any{"type": "object"}, "risk_class": "low", "safe_stop": safeStop}
	}
	return []any{
		map[string]any{"kind": "pump", "metrics": []any{map[string]any{"name": "flow", "data_type": "number", "unit": "m3/h", "writable": false}},
			"commands": []any{command("start", false), command("halt", true)}},
		map[string]any{"kind": "valve", "metrics": []any{}, "commands": []any{}},
	}
}

func stoppableBirth(t *testing.T) Record {
	t.Helper()
	record, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentHardware, ProtocolMax: 1, PriorLastSeq: -1, ClockMode: "wall"}, stoppableKinds(),
		[]model.Board{{Name: "b-1", Members: []string{"P1", "V1"}, Environment: EnvironmentHardware}})
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func interruptRig(t *testing.T) *relayRig {
	t.Helper()
	g := newRelayRig(t)
	g.relay.EnableInterrupts()
	g.relay.GatewayAttached()
	g.gw(stoppableBirth(t))
	g.relay.HostAttached()
	g.hostSends(hostState(7, 0))
	g.gw(deviceBirthOf("P1", "boot-1"))
	g.gw(deviceBirthOf("V1", "boot-7"))
	return g
}

func interruptMixedRig(t *testing.T) *mixedRig {
	t.Helper()
	m := mixedRigWith(t, simulatedBirthInbound(t, "s-1", []string{"S1", "S2"}, stoppableKinds()))
	m.relay.EnableInterrupts()
	m.relay.GatewayAttached()
	m.gw(stoppableBirth(t))
	m.relay.HostAttached()
	m.hostSends(hostState(7, 0))
	m.gw(deviceBirthOf("P1", "boot-1"))
	m.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, stoppableKinds()).Fields})
	m.simSends(simDeviceBirth("S1"))
	return m
}

func uniqueCommand(device, id string) Record {
	record := commandFor(device, id)
	record.Fields["idempotency_key"] = Command{Device: device, Name: "start", Args: map[string]any{"speed": 2}, Boot: "boot-1"}.Content()
	return record
}

func control(t *testing.T, r *Relay, line string) (map[string]any, Moves) {
	t.Helper()
	return r.Control(controlOf(t, line))
}

func eventsOf(entries []InterruptEntry) []string {
	events := make([]string, len(entries))
	for i, entry := range entries {
		events[i] = entry.Event()
	}
	return events
}

func receiptOf(t *testing.T, records []map[string]any, status, reason string) {
	t.Helper()
	if len(records) != 1 {
		t.Fatalf("up = %v, want one %s receipt of %q", records, status, reason)
	}
	given, _ := records[0]["reason"].(string)
	if records[0]["type"] != "DRECEIPT" || records[0]["status"] != status || given != reason {
		t.Fatalf("up = %v, want one %s receipt of %q", records, status, reason)
	}
}

func TestAHoldOnTheGatewaysBoardRefusesTheHostsCommandAndTheGatewayNeverSeesIt(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	reply, moves := control(t, g.relay, scopeLine(OpHold, ScopeBoard, "b-1"))
	if reply["result"] != ResultApplied || !slices.Equal(eventsOf(moves.Interrupt), []string{EventInventory, EventControl, EventReply}) {
		t.Fatalf("reply %v events %v, want the inventory, the control and the reply logged", reply, eventsOf(moves.Interrupt))
	}
	moves = g.hostSends(hostCommand())
	receiptOf(t, moves.Up, "refused", ReasonLocalControl)
	if moves.Up[0]["boot_id"] != "none" || len(moves.Down) != 0 || !slices.Equal(eventsOf(moves.Interrupt), []string{EventCommand, EventVerdict}) {
		t.Fatalf("moves = %+v, want a refusal before any boot and nothing for the gateway", moves)
	}
	control(t, g.relay, scopeLine(OpRelease, ScopeBoard, "b-1"))
	moves = g.hostSends(commandFor("P1", "c-2"))
	if !slices.Equal(typesOf(moves.Down), []string{"DCMD"}) || len(moves.Up) != 0 {
		t.Fatalf("after the release: down %v up %v, want the command carried", typesOf(moves.Down), typesOf(moves.Up))
	}
	sameContent(t, moves.Down[0], g.hostSent)
}

func TestACommandWaitsForApprovalAndGoesDownUnchangedWhenTheOperatorApprovesIt(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	control(t, g.relay, approveModeLine(ScopeDevice, "P1", true))
	moves := g.hostSends(hostCommand())
	receiptOf(t, moves.Up, "queued", "")
	if moves.Up[0]["boot_id"] != "boot-1" || len(moves.Down) != 0 {
		t.Fatalf("moves = %+v, want a queued receipt of the device's boot and nothing for the gateway", moves)
	}
	reply, approved := control(t, g.relay, `{"op":"approve","command_id":"c-1"}`)
	if reply["result"] != ResultApplied || !slices.Equal(typesOf(approved.Down), []string{"DCMD"}) || len(approved.Up) != 0 {
		t.Fatalf("reply %v down %v up %v, want the command forwarded and no further receipt from the relay", reply, typesOf(approved.Down), typesOf(approved.Up))
	}
	sameContent(t, approved.Down[0], g.hostSent)
}

func TestADeniedCommandIsRefusedAndAWaitingOneExpiresWhenItsDeadlinePasses(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	control(t, g.relay, approveModeLine(ScopeAll, "", true))
	g.hostSends(uniqueCommand("P1", "c-deny"))
	g.hostSends(uniqueCommand("V1", "c-late"))
	_, denied := control(t, g.relay, `{"op":"deny","command_id":"c-deny"}`)
	receiptOf(t, denied.Up, "refused", ReasonLocalControl)
	g.relay.SetWall(hostDeadlineNS)
	if moves := tickOnce(g.relay); len(moves.Up) != 0 {
		t.Fatalf("at the deadline up = %v, want the command to wait", moves.Up)
	}
	g.relay.SetWall(hostDeadlineNS + 1)
	expired := tickOnce(g.relay)
	receiptOf(t, expired.Up, "refused", "expired")
	if expired.Up[0]["command_id"] != "c-late" {
		t.Fatalf("receipt = %v, want it for c-late", expired.Up[0])
	}
}

func TestAnEmergencyStopSendsTheSafeStopToTheGatewayLatchesAndConsumesItsAnswers(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	reply, moves := control(t, g.relay, scopeLine(OpEStop, ScopeDevice, "P1"))
	if len(moves.Down) != 1 {
		t.Fatalf("down = %v, want the safe stop alone", typesOf(moves.Down))
	}
	stop := moves.Down[0]
	if stop["type"] != "DCMD" || stop["device"] != "P1" || stop["command"] != "halt" || stop["command_id"] != "node-stop-1" || stop["expected_boot_id"] != "boot-1" {
		t.Fatalf("stop = %v, want the declared safe stop of P1 on its boot", stop)
	}
	if reply["stopped"].([]any)[0] != "P1" {
		t.Fatalf("reply = %v, want P1 stopped", reply)
	}
	receiptOf(t, g.hostSends(hostCommand()).Up, "refused", ReasonLocalControl)
	answered := g.gw(Receipt("node-stop-1", "accepted", "", "boot-1"))
	finished := g.gw(Result("node-stop-1", "succeeded", "boot-1"))
	if len(answered.Up)+len(finished.Up) != 0 || g.relay.Stats().Forged != 0 || g.relay.Stats().Dropped != 0 {
		t.Fatalf("up %v %v stats %+v, want the relay's own answers consumed, not carried up and not counted as dropped", answered.Up, finished.Up, g.relay.Stats())
	}
	if got := eventsOf(finished.Interrupt); !slices.Equal(got, []string{EventEffect}) || finished.Interrupt[0].Text("outcome") != "succeeded" {
		t.Fatalf("effects = %v, want the result recorded as evidence", finished.Interrupt)
	}
}

func TestAHostsSafeStopIsNeverHeldByTheRelay(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	control(t, g.relay, scopeLine(OpHold, ScopeAll, ""))
	control(t, g.relay, approveModeLine(ScopeAll, "", true))
	moves := g.hostEdited(commandFor("P1", "c-stop"), func(record map[string]any) { record["command"] = "halt" })
	if !slices.Equal(typesOf(moves.Down), []string{"DCMD"}) || len(moves.Up) != 0 {
		t.Fatalf("down %v up %v, want the host's own safe stop carried through a hold", typesOf(moves.Down), typesOf(moves.Up))
	}
	receiptOf(t, g.hostSends(commandFor("P1", "c-start")).Up, "refused", ReasonLocalControl)
}

func TestWhenTheHostIsLostWhatWaitsForApprovalIsRefusedAsTheSessionEnding(t *testing.T) {
	t.Parallel()
	g := interruptRig(t)
	control(t, g.relay, approveModeLine(ScopeAll, "", true))
	g.hostSends(hostCommand())
	lost := g.relay.HostLost()
	if got := eventsOf(lost.Interrupt); !slices.Equal(got, []string{EventFlush, EventResolve}) || lost.Interrupt[1].Text("reason") != "session_ending" {
		t.Fatalf("events = %v, want a flush and the command resolved as session_ending", got)
	}
	if pending := g.relay.InterruptView()["pending"].([]any); len(pending) != 0 {
		t.Fatalf("pending = %v, want nothing left to approve", pending)
	}
}

func TestAHoldAndAnEmergencyStopActOnSimulatedDevicesOfAMixedRelayToo(t *testing.T) {
	t.Parallel()
	g := interruptMixedRig(t)
	control(t, g.relay, scopeLine(OpHold, ScopeDevice, "S1"))
	moves := g.hostSends(commandFor("S1", "c-1"))
	receiptOf(t, moves.Up, "refused", ReasonLocalControl)
	if len(moves.ToSim) != 0 || len(moves.Down) != 0 {
		t.Fatalf("to the world %v down %v, want a held command to reach neither", moves.ToSim, moves.Down)
	}
	_, stopped := control(t, g.relay, scopeLine(OpEStop, ScopeBoard, "s-1"))
	if len(stopped.ToSim) != 1 || stopped.ToSim[0]["device"] != "S1" || stopped.ToSim[0]["command"] != "halt" || stopped.ToSim[0]["expected_boot_id"] != "sim-boot" {
		t.Fatalf("to the world = %v, want the safe stop of S1 alone: S2 is not born", stopped.ToSim)
	}
	answer := g.simSends(Receipt("node-stop-1", "accepted", "", "sim-boot"))
	if len(answer.Up) != 0 || len(answer.Interrupt) != 1 || answer.Interrupt[0].Event() != EventEffect {
		t.Fatalf("answer = %+v, want the world's answer to the relay's own stop consumed and recorded", answer)
	}
}

func TestTheInventoryNamesEveryDeviceOfBothOriginsWithItsBoardAndClock(t *testing.T) {
	t.Parallel()
	g := interruptMixedRig(t)
	control(t, g.relay, scopeLine(OpHold, ScopeAll, ""))
	devices := g.relay.InterruptView()["devices"].([]any)
	got := map[string]string{}
	for _, entry := range devices {
		fields := entry.(map[string]any)
		got[fields["device"].(string)] = fields["board"].(string)
	}
	want := map[string]string{"P1": "b-1", "V1": "b-1", "S1": "s-1", "S2": "s-1"}
	if len(got) != len(want) {
		t.Fatalf("devices = %v, want %v", got, want)
	}
	for device, board := range want {
		if got[device] != board {
			t.Errorf("%s on %q, want %q", device, got[device], board)
		}
	}
}
