package domain

import "testing"

func TestTheLiveSnapshotCarriesTheInterruptStateOnlyOnceItIsSetAndOnlyChangesWithIt(t *testing.T) {
	t.Parallel()
	l := NewLive(false, nil, nil)
	if _, present := l.Fields()["interrupts"]; present {
		t.Fatal("a snapshot before any control holds interrupts, want the member absent")
	}
	if !l.SetInterrupts(plantGate().View()) || l.SetInterrupts(plantGate().View()) {
		t.Fatal("setting the same state twice must change the projection once")
	}
	g := plantGate()
	apply(t, g, scopeLine(OpHold, ScopeBoard, "b-1"))
	if !l.SetInterrupts(g.View()) {
		t.Fatal("a held board did not change the projection")
	}
	shown, _ := l.Fields()["interrupts"].(map[string]any)
	holds, _ := shown["holds"].([]any)
	devices, _ := shown["devices"].([]any)
	if len(holds) != 1 || len(devices) != 2 || devices[0].(map[string]any)["mode"] != "hold" {
		t.Fatalf("interrupts = %v, want the hold on b-1 and its two devices", shown)
	}
}

func TestACommandRefusedAsLocalControlIsShownRefusedWithThatReasonBeforeAnyBoot(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	refusal := map[string]any{"type": "DRECEIPT", "command_id": "c-1", "status": "refused", "reason": ReasonLocalControl, "boot_id": "none"}
	if !l.Observe(TraceNode, refusal, 3) {
		t.Fatal("the refusal changed nothing")
	}
	if c := l.command("c-1"); c.Receipt != "refused" || c.Reason != ReasonLocalControl || c.Report != "pending" || c.Effect != "unknown" {
		t.Fatalf("command = %+v, want it refused as local_control with nothing reported", c)
	}
}

func TestACommandThatWaitsForApprovalStaysPendingUntilTheTerminalReceipt(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	if l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "boot-1", "queued"), 3) || l.command("c-1").Receipt != "pending" {
		t.Fatalf("a queued receipt settled the command: %+v", l.command("c-1"))
	}
	if !l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "boot-1", "accepted"), 4) || l.command("c-1").Receipt != "accepted" {
		t.Fatalf("the terminal receipt did not settle the command: %+v", l.command("c-1"))
	}
}
