package device

import (
	"encoding/json"
	"os"
	"testing"
)

// validCommand loads the vendored golden command fixture (fan-01, boot-A, in
// range, deadline 7_200_000us) so device tests exercise the real contract, then
// applies optional mutations.
func validCommand(t *testing.T, mutate func(map[string]any)) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("contract/conformance/v1/valid/command.json")
	if err != nil {
		t.Fatal(err)
	}
	var cmd map[string]any
	if err := json.Unmarshal(raw, &cmd); err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(cmd)
	}
	return cmd
}

func TestAcceptedCommandEnergizesAndVerifies(t *testing.T) {
	d := New(Config{})
	out := d.ApplyCommand(validCommand(t, nil))
	if out.Receipt["accepted"] != true {
		t.Fatalf("valid command must be accepted: %v", out.Receipt)
	}
	if out.Result["status"] != "executed" {
		t.Fatalf("valid command must execute: %v", out.Result)
	}
	state := d.State()
	output, ok := state["current_output"].(map[string]any)
	if !ok || output["energized"] != true {
		t.Fatalf("fan should be energized after an accepted duty>0 command: %v", state["current_output"])
	}
	// The encoded receipt and result must themselves be valid wire records.
	if _, _, _, err := d.HandleCommandFor(t, validCommand(t, func(c map[string]any) { c["command_id"] = "cmd-2"; c["idempotency_key"] = keyB })); err != nil {
		t.Fatalf("second valid command must encode cleanly: %v", err)
	}
}

// HandleCommandFor is a small test bridge: encode the command, hand the bytes to
// HandleCommand, and surface any error. It keeps the wire path under test.
func (d *Device) HandleCommandFor(t *testing.T, command map[string]any) ([]byte, []byte, bool, error) {
	t.Helper()
	frame, err := EncodeRecord(command)
	if err != nil {
		t.Fatalf("encode command: %v", err)
	}
	return d.HandleCommand(frame)
}

const keyB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func TestRejections(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]any)
		clock  int64
		code   string
	}{
		{"wrong_boot", func(c map[string]any) { c["expected_boot_id"] = "boot-Z" }, 0, "wrong_boot"},
		{"expired", nil, 8_000_000, "expired"},
		{"out_of_range_high", func(c map[string]any) {
			c["parameters"] = map[string]any{"duty_permille": float64(900), "lease_ms": float64(5000)}
		}, 0, "out_of_range"},
		{"wrong_target", func(c map[string]any) { c["target"] = "pump-01" }, 0, "wrong_target"},
		{"unknown_operation", func(c map[string]any) { c["operation"] = "set_indicator" }, 0, "unknown_operation"},
		{"missing_param", func(c map[string]any) { c["parameters"] = map[string]any{"duty_permille": float64(450)} }, 0, "out_of_range"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mono := tc.clock
			d := New(Config{Clock: func() int64 { return mono }})
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
	d := New(Config{})
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

func TestStuckActuatorIsDesiredNotObserved(t *testing.T) {
	d := New(Config{})
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
	d := New(Config{})
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
	d := New(Config{})
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
