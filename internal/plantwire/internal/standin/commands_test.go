package standin

import (
	"reflect"
	"strings"
	"testing"
)

func dcmd(id, device, command, boot string) map[string]any {
	return map[string]any{"type": "DCMD", "lane": "control", "protocol_version": 1, "session_id": "s-1", "seq": 3, "command_id": id,
		"idempotency_key": "sha256:" + strings.Repeat("a", 64), "device": device, "command": command, "args": map[string]any{},
		"expected_boot_id": boot, "deadline_ns": 1000, "not_before_ns": 0, "authority_epoch": 1, "policy_digest": policy,
		"fallback_on_lease_loss": "safe_off"}
}

func repliesScript() Script {
	return Script{Replies: []Reply{
		{Command: "set_cooling", Receipt: "accepted", Result: "succeeded"},
		{CommandID: "c-9", Receipt: "accepted", Result: "failed"},
		{CommandID: "c-q", Receipt: "queued"},
		{CommandID: "c-u", Receipt: "accepted", Result: "unknown"},
	}}
}

func receipt(id, status, reason, boot string) map[string]any {
	record := map[string]any{"type": "DRECEIPT", "command_id": id, "status": status, "boot_id": boot}
	if reason != "" {
		record["reason"] = reason
	}
	return record
}

func result(id, outcome, boot string) map[string]any {
	return map[string]any{"type": "DRESULT", "command_id": id, "reported_outcome": outcome, "boot_id": boot}
}

func TestACommandIsAnsweredByTheBoardOrByItsScript(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command map[string]any
		want    []map[string]any
	}{
		{"an unknown device", dcmd("c-1", "pump-01", "set_cooling", "boot-7"), []map[string]any{receipt("c-1", "refused", "unknown_device", "boot-7")}},
		{"another boot", dcmd("c-1", "fan-01", "set_cooling", "boot-0"), []map[string]any{receipt("c-1", "refused", "wrong_boot", "boot-1")}},
		{"a command the kind does not have", dcmd("c-1", "fan-01", "spin", "boot-1"), []map[string]any{receipt("c-1", "refused", "unknown_command", "boot-1")}},
		{"a command nothing scripted", dcmd("c-1", "fan-01", "stop_fan", "boot-1"), []map[string]any{receipt("c-1", "refused", "not_ready", "boot-1")}},
		{"a scripted name", dcmd("c-1", "fan-01", "set_cooling", "boot-1"),
			[]map[string]any{receipt("c-1", "accepted", "", "boot-1"), result("c-1", "succeeded", "boot-1")}},
		{"a scripted id over a scripted name", dcmd("c-9", "fan-01", "set_cooling", "boot-1"),
			[]map[string]any{receipt("c-9", "accepted", "", "boot-1"), result("c-9", "failed", "boot-1")}},
		{"a scripted id without a result", dcmd("c-q", "fan-01", "stop_fan", "boot-1"), []map[string]any{receipt("c-q", "queued", "", "boot-1")}},
		{"a result nobody can vouch for", dcmd("c-u", "fan-01", "stop_fan", "boot-1"),
			[]map[string]any{receipt("c-u", "accepted", "", "boot-1"), result("c-u", "unknown", "boot-1")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := boardNode(t, repliesScript()).Handle(tc.command)
			if len(got) != len(tc.want) {
				t.Fatalf("Handle = %v, want %d records", got, len(tc.want))
			}
			for i, want := range tc.want {
				if err := wireErr(got[i]); err != nil {
					t.Errorf("record %d fails the contract: %v", i, err)
				}
				assertAnswer(t, got[i], want)
			}
		})
	}
}

func assertAnswer(t *testing.T, got, want map[string]any) {
	t.Helper()
	for key, value := range want {
		if !reflect.DeepEqual(got[key], value) {
			t.Errorf("%s = %v, want %v in %v", key, got[key], value, got)
		}
	}
	if _, reasoned := got["reason"]; reasoned && want["reason"] == nil {
		t.Errorf("record %v carries a reason the script gave none", got)
	}
}

func TestOnlyAValidCommandRecordIsAnswered(t *testing.T) {
	t.Parallel()
	node := boardNode(t, repliesScript())
	invalid := dcmd("c-1", "fan-01", "set_cooling", "boot-1")
	delete(invalid, "deadline_ns")
	release := map[string]any{"type": "HSYNC", "lane": "control", "protocol_version": 1, "session_id": "s-1", "seq": 3, "through_seq": 1}
	if got := node.Handle(invalid); got != nil {
		t.Errorf("a command that fails the contract: Handle = %v, want no answer", got)
	}
	if got := node.Handle(release); got != nil {
		t.Errorf("a record that is no command: Handle = %v, want no answer", got)
	}
}

func TestAnswersFollowTheBoardsOwnSequence(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, repliesScript()), hostOptions("hardware"))
	b.handshake()
	answered := b.node.Handle(dcmd("c-9", "fan-01", "set_cooling", "boot-1"))
	if answered[0]["seq"] != uint64(3) || answered[1]["seq"] != uint64(4) {
		t.Fatalf("answers carry seq %v and %v, want 3 and 4 right after the births", answered[0]["seq"], answered[1]["seq"])
	}
}
