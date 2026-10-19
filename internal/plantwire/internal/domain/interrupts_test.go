package domain

import (
	"fmt"
	"slices"
	"testing"
)

var plantDevices = []InterruptDevice{
	{Device: "pump-1", Board: "b-1", Clock: ClockWorld, SafeStop: "halt"},
	{Device: "pump-2", Board: "b-1", Clock: ClockWorld, SafeStop: "halt"},
	{Device: "valve-1", Board: "b-2", Clock: ClockWorld},
}

const (
	nowNS      = int64(1_000_000_000)
	deadlineNS = int64(9_000_000_000)
)

func plantGate() *InterruptGate {
	gate := NewInterruptGate()
	gate.SetInventory(plantDevices, ClockWorld, nowNS)
	return gate
}

func controlOf(t *testing.T, line string) InterruptEntry {
	t.Helper()
	request, err := DecodeControl([]byte(line))
	if err != nil {
		t.Fatalf("DecodeControl(%s) = %v", line, err)
	}
	return request
}

func scopeLine(op, kind, id string) string {
	if kind == ScopeAll {
		return fmt.Sprintf(`{"op":%q,"scope":{"kind":"all"}}`, op)
	}
	return fmt.Sprintf(`{"op":%q,"scope":{"kind":%q,"id":%q}}`, op, kind, id)
}

func approveModeLine(kind, id string, enabled bool) string {
	scope := `{"kind":"all"}`
	if kind != ScopeAll {
		scope = fmt.Sprintf(`{"kind":%q,"id":%q}`, kind, id)
	}
	return fmt.Sprintf(`{"op":"approve_mode","scope":%s,"enabled":%t}`, scope, enabled)
}

func dcmdOf(id, device string) Inbound {
	message := Message{Type: "DCMD", CommandID: id, Device: device, Command: "start", Args: map[string]any{"speed": 2},
		ExpectedBootID: "boot-1", DeadlineNS: deadlineNS}
	message.IdempotencyKey = CommandOf(message).Content()
	return Inbound{Message: message}
}

func judge(g *InterruptGate, id, device string) GateOutcome {
	return g.Judge(dcmdOf(id, device), ClockWorld, nowNS)
}

func apply(t *testing.T, g *InterruptGate, line string) GateOutcome {
	t.Helper()
	return g.Control(controlOf(t, line), ClockWorld, nowNS)
}

func verdictFor(o GateOutcome) string {
	switch {
	case o.Proceed:
		return VerdictPass
	case len(o.Queued) > 0:
		return VerdictQueue
	case len(o.Refused) > 0:
		return VerdictRefuse + ":" + o.Refused[0].Reason
	}
	return "nothing"
}

func TestAHoldRefusesEveryCommandInItsScopeAndOnlyThose(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, kind, id string
		want           map[string]string
	}{
		{"all", ScopeAll, "", map[string]string{"pump-1": "refuse:local_control", "pump-2": "refuse:local_control", "valve-1": "refuse:local_control"}},
		{"a board", ScopeBoard, "b-1", map[string]string{"pump-1": "refuse:local_control", "pump-2": "refuse:local_control", "valve-1": "pass"}},
		{"a device", ScopeDevice, "pump-2", map[string]string{"pump-1": "pass", "pump-2": "refuse:local_control", "valve-1": "pass"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := plantGate()
			if reply := apply(t, g, scopeLine(OpHold, tc.kind, tc.id)).Reply; reply["result"] != ResultApplied {
				t.Fatalf("hold = %v, want applied", reply)
			}
			for device, want := range tc.want {
				if got := verdictFor(judge(g, "c-"+device, device)); got != want {
					t.Errorf("%s: %s, want %s", device, got, want)
				}
			}
		})
	}
}

func TestAHoldWinsOverApprovalWhenBothCoverADevice(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, approveModeLine(ScopeAll, "", true))
	apply(t, g, scopeLine(OpHold, ScopeDevice, "pump-1"))
	if got := verdictFor(judge(g, "c-1", "pump-1")); got != "refuse:local_control" {
		t.Fatalf("pump-1 = %s, want the hold to win", got)
	}
	if got := verdictFor(judge(g, "c-2", "pump-2")); got != VerdictQueue {
		t.Fatalf("pump-2 = %s, want it to wait for approval", got)
	}
	devices := g.View()["devices"].([]any)
	first := devices[0].(map[string]any)
	if first["device"] != "pump-1" || first["mode"] != "hold" || devices[1].(map[string]any)["mode"] != "approve" {
		t.Fatalf("devices = %v, want pump-1 held and the others on approval", devices)
	}
}

func TestAControlThatNamesNothingTheNodeHasOrBreaksABoundIsRefusedWithItsReason(t *testing.T) {
	t.Parallel()
	g := plantGate()
	for line, reason := range map[string]string{
		scopeLine(OpHold, ScopeBoard, "b-9"):        ReasonUnknownBoard,
		scopeLine(OpEStop, ScopeDevice, "ghost"):    ReasonUnknownDevice,
		approveModeLine(ScopeDevice, "ghost", true): ReasonUnknownDevice,
		`{"op":"approve","command_id":"c-404"}`:     ReasonUnknownCommand,
		`{"op":"deny","command_id":"c-404"}`:        ReasonUnknownCommand,
	} {
		reply := apply(t, g, line).Reply
		if reply["result"] != ResultRefused || reply["reason"] != reason {
			t.Errorf("%s = %v, want refused %s", line, reply, reason)
		}
	}
	if len(g.View()["holds"].([]any)) != 0 {
		t.Fatalf("holds = %v, want a refused control to change nothing", g.View()["holds"])
	}
}

func TestTheNumberOfScopesIsBoundedAndARepeatedScopeIsNotANewOne(t *testing.T) {
	t.Parallel()
	var devices []InterruptDevice
	for i := range maxInterruptScopes + 1 {
		devices = append(devices, InterruptDevice{Device: fmt.Sprintf("d-%03d", i), Board: "b", Clock: ClockWorld})
	}
	g := NewInterruptGate()
	g.SetInventory(devices, ClockWorld, nowNS)
	for i := range maxInterruptScopes {
		if reply := apply(t, g, scopeLine(OpHold, ScopeDevice, devices[i].Device)).Reply; reply["result"] != ResultApplied {
			t.Fatalf("hold %d = %v", i, reply)
		}
	}
	over := apply(t, g, scopeLine(OpHold, ScopeDevice, devices[maxInterruptScopes].Device)).Reply
	again := apply(t, g, scopeLine(OpHold, ScopeDevice, devices[0].Device)).Reply
	if over["reason"] != ReasonTooManyScopes || again["result"] != ResultApplied {
		t.Fatalf("over %v, again %v, want the one over the bound refused and a repeat applied", over, again)
	}
}

func TestSettingAndClearingAreIdempotent(t *testing.T) {
	t.Parallel()
	g := plantGate()
	for _, line := range []string{scopeLine(OpHold, ScopeBoard, "b-1"), approveModeLine(ScopeAll, "", true)} {
		first, second := apply(t, g, line).Reply["state"], apply(t, g, line).Reply["state"]
		if !sameEntry(first.(map[string]any), second.(map[string]any)) {
			t.Fatalf("%s twice changed the state: %v then %v", line, first, second)
		}
	}
	for _, line := range []string{scopeLine(OpRelease, ScopeBoard, "b-1"), approveModeLine(ScopeAll, "", false)} {
		apply(t, g, line)
		if reply := apply(t, g, line).Reply; reply["result"] != ResultApplied {
			t.Fatalf("%s a second time = %v, want applied", line, reply)
		}
	}
	state := g.View()
	if len(state["holds"].([]any))+len(state["approve"].([]any)) != 0 {
		t.Fatalf("state = %v, want nothing set", state)
	}
}

func TestAnEmergencyStopOrdersTheSafeStopOfEachDeviceInScopeAndLatchesAHold(t *testing.T) {
	t.Parallel()
	g := plantGate()
	outcome := apply(t, g, scopeLine(OpEStop, ScopeAll, ""))
	if want := []StopOrder{{"pump-1", "halt"}, {"pump-2", "halt"}}; !slices.Equal(outcome.Stops, want) {
		t.Fatalf("stops = %v, want %v", outcome.Stops, want)
	}
	if outcome.Reply["stopped"].([]any)[1] != "pump-2" || outcome.Reply["skipped"].([]any)[0] != "valve-1" {
		t.Fatalf("reply = %v, want pump-1 and pump-2 stopped and valve-1 skipped for declaring no safe stop", outcome.Reply)
	}
	if got := verdictFor(judge(g, "c-1", "pump-1")); got != "refuse:local_control" {
		t.Fatalf("a command after the stop = %s, want it refused: the stop latches a hold", got)
	}
	hold := g.View()["holds"].([]any)[0].(map[string]any)
	if hold["latched"] != true {
		t.Fatalf("hold = %v, want it latched", hold)
	}
}

func TestOnlyReleasingTheScopeClearsALatchAndAnotherHoldDoesNot(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, scopeLine(OpEStop, ScopeBoard, "b-1"))
	apply(t, g, scopeLine(OpHold, ScopeBoard, "b-1"))
	if hold := g.View()["holds"].([]any)[0].(map[string]any); hold["latched"] != true {
		t.Fatalf("hold = %v, want a plain hold on a latched scope to leave the latch", hold)
	}
	apply(t, g, scopeLine(OpRelease, ScopeDevice, "pump-1"))
	if len(g.View()["holds"].([]any)) != 1 {
		t.Fatalf("holds = %v, want a release of another scope to clear nothing", g.View()["holds"])
	}
	apply(t, g, scopeLine(OpRelease, ScopeBoard, "b-1"))
	if got := verdictFor(judge(g, "c-1", "pump-1")); got != VerdictPass {
		t.Fatalf("after the release = %s, want the command to pass", got)
	}
}

func haltOf(id string, args map[string]any) Inbound {
	in := dcmdOf(id, "pump-1")
	in.Command, in.Args = "halt", args
	in.IdempotencyKey = CommandOf(in.Message).Content()
	return in
}

func TestTheDeclaredSafeStopWithoutArgumentsIsNeverHeldNorQueued(t *testing.T) {
	t.Parallel()
	g := plantGate()
	apply(t, g, scopeLine(OpHold, ScopeAll, ""))
	apply(t, g, approveModeLine(ScopeAll, "", true))
	outcome := g.Judge(haltOf("c-stop", map[string]any{}), ClockWorld, nowNS)
	if !outcome.Proceed || len(outcome.Refused)+len(outcome.Queued) != 0 {
		t.Fatalf("outcome = %+v, want the safe stop to pass a hold and an approval mode", outcome)
	}
}
