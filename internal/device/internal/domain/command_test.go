package domain

import (
	"strings"
	"testing"
)

func TestRejections(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]any)
		clock  int64
		code   string
	}{
		{"wrong_boot", func(c map[string]any) { c["expected_boot_id"] = "boot-Z" }, 0, "wrong_boot"},
		{"expired", func(c map[string]any) { c["not_before_mono_us"] = float64(1) }, 8_000_000, "expired"},
		{"out_of_range_high", func(c map[string]any) {
			c["parameters"] = map[string]any{"duty_permille": float64(900), "lease_ms": float64(5000)}
		}, 0, "out_of_range"},
		{"wrong_target", func(c map[string]any) { c["target"] = "pump-01" }, 0, "wrong_target"},
		{"unknown_operation", func(c map[string]any) { c["operation"] = "set_indicator" }, 0, "unknown_operation"},
		{"missing_param", func(c map[string]any) { c["parameters"] = map[string]any{"duty_permille": float64(450)} }, 0, "out_of_range"},
		{"unknown_nonnumeric_param", func(c map[string]any) {
			c["parameters"] = map[string]any{"duty_permille": float64(450), "lease_ms": float64(5000), "extra": "ignored"}
		}, 0, "out_of_range"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mono := tc.clock
			d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return mono }})
			out := d.ApplyCommand(validCommand(t, tc.mutate))
			if out.Receipt["accepted"] != false {
				t.Fatalf("%s must be rejected: %v", tc.name, out.Receipt)
			}
			if out.Receipt["reject_code"] != tc.code {
				t.Fatalf("%s expected reject_code %q, got %v", tc.name, tc.code, out.Receipt["reject_code"])
			}
			if out.Result["status"] != "rejected" {
				t.Fatalf("%s result must be rejected: %v", tc.name, out.Result)
			}
			if state := d.State(); state["current_output"] != nil {
				t.Fatalf("%s must not energize any output: %v", tc.name, state["current_output"])
			}
		})
	}
}

func TestIdempotentReplayAppliesNoSecondEffect(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	first := d.ApplyCommand(validCommand(t, nil))
	second := d.ApplyCommand(validCommand(t, nil))
	if first.Receipt["command_id"] != second.Receipt["command_id"] {
		t.Fatal("replay must return the prior receipt")
	}
	state := d.State()
	ledger := state["dedup_ledger"].(map[string]any)
	if ledger["size"] != float64(1) {
		t.Fatalf("a replayed key must be deduped to a single ledger entry: %v", ledger["size"])
	}
}

func TestIdempotencyKeyConflictIsRejected(t *testing.T) {
	plant := &countingPlant{}
	d := New(Config{Capabilities: testCaps(t), Plant: plant})
	first := validCommand(t, nil)
	if out := d.ApplyCommand(first); out.Receipt["accepted"] != true {
		t.Fatalf("first command must be accepted: %v", out.Receipt)
	}
	conflict := validCommand(t, func(c map[string]any) {
		c["command_id"] = "cmd-conflict"
		c["parameters"] = map[string]any{"duty_permille": float64(600), "lease_ms": float64(5000)}
	})
	out := d.ApplyCommand(conflict)
	if out.Receipt["accepted"] != false || out.Receipt["reject_code"] != "duplicate" || out.Result["error_code"] != "duplicate" {
		t.Fatalf("conflicting idempotency reuse must be rejected: receipt=%v result=%v", out.Receipt, out.Result)
	}
	if plant.applyCalls != 1 {
		t.Fatalf("conflicting idempotency reuse must not apply a second effect: %d", plant.applyCalls)
	}
}

func TestSafeStopAckLostAndRebootWithholdTerminalExchange(t *testing.T) {
	d := New(Config{
		Capabilities:  testCaps(t),
		FaultSchedule: []FaultInjection{{Name: FaultAckLost, AcceptedCommand: 1}},
	})
	safeStop := validCommand(t, func(c map[string]any) {
		c["command_id"] = "safe-stop/fan-01"
		c["idempotency_key"] = "sha256:" + strings.Repeat("e", 64)
		c["operation"] = "safe_stop"
		c["parameters"] = map[string]any{}
		c["expires_after_ms"] = float64(1000)
	})
	if out := d.ApplyCommand(safeStop); !out.AckLost || out.Result["status"] != "safe_state" {
		t.Fatalf("safe-stop ack loss must withhold its pair: %+v", out)
	}

	reboot := New(Config{
		Capabilities:  testCaps(t),
		FaultSchedule: []FaultInjection{{Name: FaultReboot, AcceptedCommand: 1}},
	})
	out := reboot.ApplyCommand(validCommand(t, nil))
	if !out.AckLost || out.Receipt["boot_id"] == reboot.BootID() {
		t.Fatalf("reboot must withhold the old-boot pair: outcome=%+v current_boot=%s", out, reboot.BootID())
	}
}

func TestStuckActuatorIsDesiredNotObserved(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	d.SetFaults(Faults{Stuck: true})
	out := d.ApplyCommand(validCommand(t, nil))
	if out.Receipt["accepted"] != true || out.Result["status"] != "executed" {
		t.Fatalf("stuck actuator is still accepted and executed at the wire level: %v %v", out.Receipt, out.Result)
	}
	output := d.State()["current_output"].(map[string]any)
	if output["energized"] != false {
		t.Fatal("a stuck actuator must report NOT energized — the desired≠observed case")
	}
}

func TestAckLostAppliesEffectButWithholdsReceipt(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	d.SetFaults(Faults{AckLost: true})
	out := d.ApplyCommand(validCommand(t, nil))
	if !out.AckLost {
		t.Fatal("ack_lost must flag the receipt as withheld from the wire")
	}
	output := d.State()["current_output"].(map[string]any)
	if output["energized"] != true {
		t.Fatal("ack_lost still applies the physical effect — the upstream sees unknown, not failure")
	}
}

func TestRebootChangesBootAndInvalidatesOldCommands(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	d.ApplyCommand(validCommand(t, nil))
	d.Reboot("boot-B")
	if d.BootID() != "boot-B" {
		t.Fatalf("reboot must change boot id, got %s", d.BootID())
	}
	if d.State()["current_output"] != nil {
		t.Fatal("reboot must drop to safe state (no energized output)")
	}
	// A command still bound to the old boot must now be rejected wrong_boot.
	out := d.ApplyCommand(validCommand(t, func(c map[string]any) { c["command_id"] = "cmd-old" }))
	if out.Receipt["reject_code"] != "wrong_boot" {
		t.Fatalf("post-reboot, an old-boot command must be wrong_boot: %v", out.Receipt)
	}
}
