package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestASafeStopNameWithArgumentsIsNotTheDeclaredStopAndPassesNoHold(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, scopeLine(OpHold, ScopeAll, ""))
	if got := verdictFor(g.Judge(haltOf("c-run", map[string]any{"mode": "run"}), ClockWorld, nowNS)); got != "refuse:local_control" {
		t.Fatalf("a halt with arguments = %s, want it held: only the declared stop without arguments is exempt", got)
	}
	big := haltOf("c-big", map[string]any{"note": strings.Repeat("x", 3000)})
	if got := verdictFor(g.Judge(big, ClockWorld, nowNS)); got != "refuse:local_control" {
		t.Fatalf("a halt whose arguments were left out of the log = %s, want it held", got)
	}
}

func TestTheSafeStopExemptionIsDecidedByTheMachineAndNotByAFlagInTheLoggedInput(t *testing.T) {
	t.Parallel()
	machine := NewInterrupts()
	machine.Step(newEntry(EventInventory, "devices", []any{plantDevices[0].Fields()}).at(ClockWorld, nowNS))
	machine.Step(newEntry(EventControl, "op", OpHold, "scope", Scope{Kind: ScopeAll}.Fields()).at(ClockWorld, nowNS))
	forged := newEntry(EventCommand, "command_id", "c-1", "key", "k", "content", "d", "device", "pump-1", "name", "start", "deadline_ns", deadlineNS,
		"args", map[string]any{}, "args_truncated", false, "safe_stop", true).at(ClockWorld, nowNS)
	if out := machine.Step(forged); len(out) != 1 || out[0].Text("verdict") != VerdictRefuse || out[0].Text("reason") != ReasonLocalControl {
		t.Fatalf("decision = %v, want a command that claims safe_stop but is not the declared stop refused", out)
	}
}

func TestACommandIdThatIsAlreadyWaitingIsRefusedAsADuplicateAndNeverForwardsAnotherPayload(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	judge(g, "c-1", "pump-1")
	if got := verdictFor(judge(g, "c-1", "pump-2")); got != "refuse:duplicate" {
		t.Fatalf("a second command under the waiting id = %s, want a duplicate", got)
	}
	repeat := dcmdOf("c-1", "pump-1")
	if got := verdictFor(g.Judge(repeat, ClockWorld, nowNS)); got != "refuse:duplicate" {
		t.Fatalf("the same id again = %s, want a duplicate: an id names one waiting command", got)
	}
	approved := apply(t, g, `{"op":"approve","command_id":"c-1"}`)
	if len(approved.Forward) != 1 || approved.Forward[0].Device != "pump-1" {
		t.Fatalf("forward = %+v, want the command that waited under c-1", approved.Forward)
	}
}

func TestAWaitingCommandShowsItsArgumentsAndDigestAndAnApprovalMustNameWhatItWasShown(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	waiting := dcmdOf("c-1", "pump-1")
	g.Judge(waiting, ClockWorld, nowNS)
	entry := g.View()["pending"].([]any)[0].(map[string]any)
	digest := CommandOf(waiting.Message).Content()
	if entry["content_digest"] != digest || entry["args_truncated"] != false || entry["args"].(map[string]any)["speed"] != 2 {
		t.Fatalf("pending = %v, want the digest %s and the arguments the command carries", entry, digest)
	}
	for _, op := range []string{OpApprove, OpDeny} {
		line := fmt.Sprintf(`{"op":%q,"command_id":"c-1","content_digest":"sha256:%s"}`, op, strings.Repeat("0", 64))
		if reply := apply(t, g, line).Reply; reply["result"] != ResultRefused || reply["reason"] != ReasonUnknownCommand {
			t.Fatalf("%s of other content = %v, want refused unknown_command", op, reply)
		}
	}
	if len(g.View()["pending"].([]any)) != 1 {
		t.Fatal("a refused approval or denial changed what waits")
	}
	line := fmt.Sprintf(`{"op":"approve","command_id":"c-1","content_digest":%q}`, digest)
	if approved := apply(t, g, line); len(approved.Forward) != 1 || approved.Reply["result"] != ResultApplied {
		t.Fatalf("approval naming the shown digest = %+v, want it forwarded", approved)
	}
}

func TestArgumentsTooLargeToShowAreLeftOutAndSaidSo(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	big := dcmdOf("c-1", "pump-1")
	big.Args = map[string]any{"note": strings.Repeat("x", maxLoggedArgsBytes+1)}
	big.IdempotencyKey = CommandOf(big.Message).Content()
	outcome := g.Judge(big, ClockWorld, nowNS)
	logged := outcome.Entries[0]
	entry := g.View()["pending"].([]any)[0].(map[string]any)
	if logged["args_truncated"] != true || len(logged["args"].(map[string]any)) != 0 || entry["args_truncated"] != true || len(entry["args"].(map[string]any)) != 0 {
		t.Fatalf("logged %v pending %v, want the arguments left out of both with args_truncated true", logged, entry)
	}
}

func TestAnEmergencyStopIsNeverRefusedForCapacityAndMakesRoomByDroppingWhatItCovers(t *testing.T) {
	t.Parallel()
	var devices []InterruptDevice
	for i := range maxInterruptScopes + 2 {
		devices = append(devices, InterruptDevice{Device: fmt.Sprintf("d-%03d", i), Board: "b", Clock: ClockWorld, SafeStop: "halt"})
	}
	g := NewInterruptGate()
	g.SetInventory(devices, ClockWorld, nowNS)
	for i := range maxInterruptScopes {
		apply(t, g, approveModeLine(ScopeDevice, devices[i].Device, true))
	}
	if reply := apply(t, g, scopeLine(OpHold, ScopeDevice, devices[maxInterruptScopes].Device)).Reply; reply["reason"] != ReasonTooManyScopes {
		t.Fatalf("a hold over the bound = %v, want too_many_scopes", reply)
	}
	stop := apply(t, g, scopeLine(OpEStop, ScopeBoard, "b"))
	if stop.Reply["result"] != ResultApplied || len(stop.Stops) != len(devices) {
		t.Fatalf("estop = %v with %d stops, want it applied for all %d devices", stop.Reply, len(stop.Stops), len(devices))
	}
	state := g.View()
	holds, approves := state["holds"].([]any), state["approve"].([]any)
	if len(holds) != 1 || holds[0].(map[string]any)["latched"] != true || len(holds)+len(approves) != maxInterruptScopes {
		t.Fatalf("state = %v, want the latch kept and just enough covered approval scopes dropped to stay at %d", state, maxInterruptScopes)
	}
}

func TestAnEmergencyStopOverLatchesItCannotDropIsStillAppliedAndLatched(t *testing.T) {
	t.Parallel()
	var devices []InterruptDevice
	for i := range maxInterruptScopes + 1 {
		devices = append(devices, InterruptDevice{Device: fmt.Sprintf("d-%03d", i), Board: fmt.Sprintf("b-%03d", i), Clock: ClockWorld, SafeStop: "halt"})
	}
	g := NewInterruptGate()
	g.SetInventory(devices, ClockWorld, nowNS)
	for _, device := range devices {
		if reply := apply(t, g, scopeLine(OpEStop, ScopeDevice, device.Device)).Reply; reply["result"] != ResultApplied {
			t.Fatalf("estop of %s = %v, want it applied however many latches exist", device.Device, reply)
		}
	}
	if holds := g.View()["holds"].([]any); len(holds) != len(devices) {
		t.Fatalf("holds = %d, want %d latches", len(holds), len(devices))
	}
}

func TestADeviceTheNodeCannotReachIsReportedUnreachableAndNeverAsStopped(t *testing.T) {
	t.Parallel()
	devices := []InterruptDevice{{Device: "P1", Board: "b", Clock: ClockWall, SafeStop: "halt"},
		{Device: "P2", Board: "b", Clock: ClockWall, SafeStop: "halt", Unreachable: true}, {Device: "P3", Board: "b", Clock: ClockWall}}
	g := NewInterruptGate()
	g.SetInventory(devices, ClockWall, nowNS)
	stop := apply(t, g, scopeLine(OpEStop, ScopeBoard, "b"))
	if !slices.Equal(stop.Stops, []StopOrder{{"P1", "halt"}}) {
		t.Fatalf("stops = %v, want an order for P1 alone", stop.Stops)
	}
	for member, want := range map[string]string{"stopped": "P1", "unreachable": "P2", "skipped": "P3"} {
		if listed := stop.Reply[member].([]any); len(listed) != 1 || listed[0] != want {
			t.Errorf("%s = %v, want [%s]", member, listed, want)
		}
	}
}

func TestAnEffectNoStopOrderedAndAPurgeNoCommandPassedAreNotEvidence(t *testing.T) {
	t.Parallel()
	for name, record := range map[string]InterruptEntry{
		"an effect of a device nothing stopped": EffectEntry("node-stop-1", "pump-1", "accepted", ClockWall, nowNS),
		"a purge of a command nothing passed":   PurgeEntry("c-9", "pump-1", ReasonLocalControl, ClockWorld, nowNS),
	} {
		err := ReplayInterrupts([]map[string]any{parsedRecord(t, record)})
		if !errors.Is(err, ErrInterruptReplay) || !strings.Contains(err.Error(), "record 0") {
			t.Errorf("%s: ReplayInterrupts = %v, want ErrInterruptReplay naming record 0", name, err)
		}
	}
}
