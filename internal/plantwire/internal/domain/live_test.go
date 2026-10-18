package domain

import (
	"encoding/json"
	"maps"
	"testing"
)

func liveRecord(record Record) map[string]any {
	r := maps.Clone(record.Fields)
	r["type"] = record.Type
	return r
}

func liveBoard(t *testing.T) *Live {
	t.Helper()
	l := NewLive(true, nil, []Bound{{Instance: "pump-1", Board: "b-1", Device: "P1", Kind: "pump"}})
	l.Observe(TraceNode, liveRecord(hardwareBirth(t, nil)), 0)
	l.kinds["pump"]["flow"]["eu_range"] = map[string]any{"min": 0.0, "max": 120.0}
	l.kinds["pump"]["flow"]["scan_class"] = "1s"
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-1")), 1)
	return l
}

func liveAnswer(kind, id, boot, status string) map[string]any {
	field := "status"
	if kind == "DRESULT" {
		field = "reported_outcome"
	}
	return map[string]any{"type": kind, "command_id": id, "boot_id": boot, field: status}
}

func acceptedCommand(t *testing.T, l *Live) {
	t.Helper()
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "boot-1", "accepted"), 3)
	l.Observe(TraceNode, liveAnswer("DRESULT", "c-1", "boot-1", "succeeded"), 4)
}

func TestLiveKeepsAcknowledgementBoardReportAndHostEvidenceSeparate(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	c := l.command("c-1")
	if c.Receipt != "pending" || c.Report != "pending" || c.Effect != "unknown" {
		t.Fatalf("sent = %+v", c)
	}
	l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "boot-1", "accepted"), 3)
	if c.Receipt != "accepted" || c.Report != "pending" || c.Effect != "unknown" {
		t.Fatalf("acknowledged = %+v", c)
	}
	l.Observe(TraceNode, liveAnswer("DRESULT", "c-1", "boot-1", "succeeded"), 4)
	if c.Report != "succeeded" || c.Effect != "unknown" {
		t.Fatalf("self-report = %+v", c)
	}
	if !l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 5) || c.Effect != "observed" {
		t.Fatalf("host independent evidence = %+v", c)
	}
	if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 6) || c.UpdatedAtNS != 5 {
		t.Fatalf("duplicate evidence changed terminal decision: %+v", c)
	}
}

func hostEvidence() map[string]any {
	return map[string]any{"command_id": "c-1", "device": "P1", "boot_id": "boot-1", "metric": "flow", "outcome": "observed",
		"expected": 3.5, "tolerance": 0.1, "observed_value": 3.5}
}

func TestLiveRefusesMismatchedAndPrematureHostEvidence(t *testing.T) {
	t.Parallel()
	for _, member := range []string{"command_id", "device", "boot_id", "outcome"} {
		t.Run(member, func(t *testing.T) {
			t.Parallel()
			l := liveBoard(t)
			acceptedCommand(t, l)
			e := hostEvidence()
			e[member] = "other"
			if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{e}}, 5) || l.command("c-1").Effect != "unknown" {
				t.Fatalf("mismatched %s confirmed effect", member)
			}
		})
	}
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 3) {
		t.Fatal("evidence before acceptance and successful report confirmed")
	}
}

func TestLiveShowsRefusalAndIgnoresWrongBootAndRegressingAnswers(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	if l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "wrong", "accepted"), 3) {
		t.Fatal("wrong boot accepted")
	}
	refused := liveAnswer("DRECEIPT", "c-1", "boot-1", "refused")
	refused["reason"] = "not_ready"
	l.Observe(TraceNode, refused, 4)
	l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "boot-1", "accepted"), 5)
	l.Observe(TraceNode, liveAnswer("DRESULT", "c-1", "boot-1", "succeeded"), 6)
	c := l.command("c-1")
	if c.Receipt != "refused" || c.Reason != "not_ready" || c.Report != "pending" || c.Effect != "unknown" {
		t.Fatalf("refused command regressed: %+v", c)
	}
}

func TestLiveReadingsFlagBoundsStalenessAndDisconnectAndResetOnReboot(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	d := l.devices["P1"]
	if d.Instance != "pump-1" || d.Environment != EnvironmentHardware || d.Readings[0].State != "good" {
		t.Fatalf("bound reading = %+v", d)
	}
	l.Tick(3000)
	if d.Readings[0].State != "stale" {
		t.Fatalf("silent metric = %+v", d.Readings)
	}
	data := map[string]any{"type": "DDATA", "device": "P1", "boot_id": "boot-1", "metrics": []any{
		map[string]any{"alias": 1, "value": -2.0, "quality": "good"}}}
	l.Observe(TraceNode, data, 5)
	if d.Readings[0].State != "invalid" {
		t.Fatalf("negative flow shown normal: %+v", d.Readings)
	}
	acceptedCommand(t, l)
	l.Connection("gateway", "offline", 6)
	if d.Health != "offline" || d.Readings[0].State != "disconnected" {
		t.Fatalf("lost device = %+v", d)
	}
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-2")), 7)
	if d.Reboots != 1 || d.BootID != "boot-2" || d.Readings[0].State != "good" {
		t.Fatalf("reboot = %+v", d)
	}
	if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 8) {
		t.Fatal("old boot's evidence confirmed after reboot")
	}
}

func TestLiveProjectionReplaysIdenticallyFromTheSameAdmittedRecords(t *testing.T) {
	t.Parallel()
	a, b := liveBoard(t), liveBoard(t)
	for _, l := range []*Live{a, b} {
		acceptedCommand(t, l)
		l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 5)
	}
	left, leftErr := json.Marshal(a.Fields())
	right, rightErr := json.Marshal(b.Fields())
	if leftErr != nil || rightErr != nil || string(left) != string(right) {
		t.Fatalf("record replay differs: %s %v; %s %v", left, leftErr, right, rightErr)
	}
}
