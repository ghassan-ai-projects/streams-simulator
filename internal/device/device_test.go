package device

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

type countingPlant struct {
	applyCalls    int
	safeStopCalls int
	safeStopErr   error
}

type energizedSafeStopPlant struct{}

func (energizedSafeStopPlant) Apply(PlantCommand) (PlantEffect, error) {
	return PlantEffect{Value: 450, Energized: true}, nil
}

func (energizedSafeStopPlant) SafeStop(string, int64) (PlantEffect, error) {
	return PlantEffect{Value: 450, Energized: true}, nil
}

func (p *countingPlant) Apply(PlantCommand) (PlantEffect, error) {
	p.applyCalls++
	return PlantEffect{Value: 450, Energized: true}, nil
}

func (p *countingPlant) SafeStop(string, int64) (PlantEffect, error) {
	p.safeStopCalls++
	return PlantEffect{}, p.safeStopErr
}

// testCaps loads the device capability catalog from the JSON data fixture — the
// same data-not-code path the emulator uses at runtime.
func testCaps(t *testing.T) *Capabilities {
	t.Helper()
	data, err := os.ReadFile("testdata/thermal_capability_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	caps, err := LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}

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
	// The emulator's manual clock starts at zero. The explicit freshness test
	// covers future-bound commands.
	cmd["not_before_mono_us"] = float64(0)
	if mutate != nil {
		mutate(cmd)
	}
	return cmd
}

func TestRejectsCommandBeforeNotBefore(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return 0 }})
	out := d.ApplyCommand(validCommand(t, func(c map[string]any) {
		c["not_before_mono_us"] = float64(1)
	}))
	if out.Receipt["accepted"] != false || out.Receipt["reject_code"] != "not_ready" {
		t.Fatalf("future-bound command must be rejected as not_ready: %v", out.Receipt)
	}
}

func TestLeaseExpiryReturnsSafeState(t *testing.T) {
	now := int64(0)
	d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return now }})
	out := d.ApplyCommand(validCommand(t, nil))
	if out.Receipt["accepted"] != true {
		t.Fatalf("valid command must be accepted: %v", out.Receipt)
	}
	now = 5_000_000
	state := d.State()
	if state["safe_state"] != true {
		t.Fatalf("expired lease must put the device in safe state: %v", state)
	}
	output := state["current_output"].(map[string]any)
	if output["energized"] != false {
		t.Fatalf("expired lease must de-energize the output: %v", output)
	}
}

func TestSafeStopIsAnExplicitDeviceOperation(t *testing.T) {
	plant := &countingPlant{}
	d := New(Config{Capabilities: testCaps(t), Plant: plant, Clock: func() int64 { return 0 }})
	out := d.ApplyCommand(validCommand(t, func(command map[string]any) {
		command["command_id"] = "safe-stop/fan-01"
		command["idempotency_key"] = "sha256:" + strings.Repeat("e", 64)
		command["operation"] = "safe_stop"
		command["parameters"] = map[string]any{}
		command["expires_after_ms"] = float64(1000)
	}))
	if out.Receipt["accepted"] != true || out.Result["status"] != "safe_state" {
		t.Fatalf("explicit safe stop must be accepted as safe_state: receipt=%v result=%v", out.Receipt, out.Result)
	}
	if plant.applyCalls != 0 || plant.safeStopCalls != 1 {
		t.Fatalf("safe stop must bypass ordinary plant apply: apply=%d safe_stop=%d", plant.applyCalls, plant.safeStopCalls)
	}
}

func TestSafeStopFailureDoesNotClaimSafe(t *testing.T) {
	for _, reboot := range []bool{false, true} {
		name := "lease expiry"
		if reboot {
			name = "reboot"
		}
		t.Run(name, func(t *testing.T) {
			now := int64(0)
			plant := &countingPlant{safeStopErr: errors.New("stop unavailable")}
			d := New(Config{
				Capabilities: testCaps(t),
				Plant:        plant,
				Clock:        func() int64 { return now },
			})
			if out := d.ApplyCommand(validCommand(t, nil)); out.Receipt["accepted"] != true {
				t.Fatalf("valid lease command must be accepted: %v", out.Receipt)
			}
			if reboot {
				d.Reboot("boot-failed-stop")
			} else {
				now = 5_000_000
				_ = d.State()
			}
			state := d.State()
			if state["safe_state"] != false {
				t.Fatalf("failed safe stop must not claim safe: %v", state)
			}
			output, ok := state["current_output"].(map[string]any)
			if !ok || output["energized"] != true {
				t.Fatalf("failed safe stop must preserve conservative energized evidence: %v", state["current_output"])
			}
			if plant.safeStopCalls != 1 {
				t.Fatalf("safe stop must be attempted once, got %d", plant.safeStopCalls)
			}
		})
	}
}

func TestSafeStopPositivePlantEffectDoesNotClaimSafe(t *testing.T) {
	now := int64(0)
	d := New(Config{
		Capabilities: testCaps(t),
		Plant:        energizedSafeStopPlant{},
		Clock:        func() int64 { return now },
	})
	if out := d.ApplyCommand(validCommand(t, nil)); out.Receipt["accepted"] != true {
		t.Fatalf("valid lease command must be accepted: %v", out.Receipt)
	}
	now = 5_000_000
	state := d.State()
	if state["safe_state"] != false {
		t.Fatalf("positive safe-stop plant effect must not claim safe: %v", state)
	}
}

func TestCapabilityDigestIsBoundToLoadedData(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	first, err := LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	changedData := []byte(strings.Replace(string(data), "fan-01", "fan-02", 1))
	second, err := LoadCapabilities(changedData)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest() == second.Digest() {
		t.Fatalf("different capability catalogs must have different digests: %s", first.Digest())
	}
	if got := New(Config{Capabilities: first}).State()["capability_digest"]; got != first.Digest() {
		t.Fatalf("device state must advertise the loaded catalog digest: %v vs %s", got, first.Digest())
	}
}

func TestCanonicalRouteExpiryBoundsCommandFreshness(t *testing.T) {
	tests := []struct {
		name           string
		expiresAfter   float64
		wantAccepted   bool
		wantRejectCode string
	}{
		{name: "within route", expiresAfter: 20000, wantAccepted: true},
		{name: "beyond route", expiresAfter: 20001, wantRejectCode: "expired"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plant := &countingPlant{}
			d := New(Config{Capabilities: testCaps(t), Plant: plant})
			out := d.ApplyCommand(validCommand(t, func(command map[string]any) {
				command["expires_after_ms"] = tc.expiresAfter
			}))
			if got := out.Receipt["accepted"] == true; got != tc.wantAccepted {
				t.Fatalf("accepted = %v, want %v: %v", got, tc.wantAccepted, out.Receipt)
			}
			if tc.wantRejectCode != "" && out.Receipt["reject_code"] != tc.wantRejectCode {
				t.Fatalf("reject_code = %v, want %q", out.Receipt["reject_code"], tc.wantRejectCode)
			}
			wantCalls := 0
			if tc.wantAccepted {
				wantCalls = 1
			}
			if plant.applyCalls != wantCalls {
				t.Fatalf("plant calls = %d, want %d", plant.applyCalls, wantCalls)
			}
		})
	}
}

func TestAcceptedCommandEnergizesAndVerifies(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
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

func TestCanonicalRouteAcceptsCatalogDefinedStringPresetParameter(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	out := d.ApplyCommand(validCommand(t, func(command map[string]any) {
		command["command_id"] = "led-command"
		command["idempotency_key"] = "sha256:" + strings.Repeat("f", 64)
		command["target"] = "led-01"
		command["operation"] = "set_led"
		command["parameters"] = map[string]any{
			"brightness_permille": float64(1000),
			"pattern":             "solid",
		}
		command["expires_after_ms"] = float64(60000)
	}))
	if out.Receipt["accepted"] != true || out.Result["status"] != "executed" {
		t.Fatalf("catalog-defined LED preset must be accepted: receipt=%v result=%v", out.Receipt, out.Result)
	}
	output, ok := d.State()["current_output"].(map[string]any)
	if !ok || output["operation"] != "set_led" || output["energized"] != true {
		t.Fatalf("accepted LED preset must energize the declared output: %v", output)
	}

	rejected := d.ApplyCommand(validCommand(t, func(command map[string]any) {
		command["command_id"] = "led-invalid-pattern"
		command["idempotency_key"] = "sha256:" + strings.Repeat("a", 64)
		command["target"] = "led-01"
		command["operation"] = "set_led"
		command["parameters"] = map[string]any{
			"brightness_permille": float64(1000),
			"pattern":             "unknown",
		}
		command["expires_after_ms"] = float64(60000)
	}))
	if rejected.Receipt["accepted"] != false || rejected.Receipt["reject_code"] != "out_of_range" {
		t.Fatalf("unknown catalog string preset must be rejected: %v", rejected.Receipt)
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

func TestScheduledFaultsAreDeterministic(t *testing.T) {
	cases := []struct {
		name        string
		fault       string
		wantCode    string
		wantAckLost bool
		wantApply   int
		wantBoot    string
	}{
		{name: FaultStuck, fault: FaultStuck, wantApply: 0},
		{name: FaultAckLost, fault: FaultAckLost, wantAckLost: true, wantApply: 1},
		{name: FaultExpired, fault: FaultExpired, wantCode: FaultExpired, wantApply: 0},
		{name: FaultReboot, fault: FaultReboot, wantAckLost: true, wantApply: 1, wantBoot: "boot-reboot-1"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			plant := &countingPlant{}
			d := New(Config{
				Capabilities:  testCaps(t),
				Plant:         plant,
				FaultSchedule: []FaultInjection{{Name: tc.fault, AcceptedCommand: 1}},
			})
			out := d.ApplyCommand(validCommand(t, nil))
			if out.AckLost != tc.wantAckLost {
				t.Fatalf("ack_lost = %v, want %v: %+v", out.AckLost, tc.wantAckLost, out)
			}
			if tc.wantCode == "" && out.Receipt["accepted"] != true {
				t.Fatalf("fault %q should preserve acceptance: %v", tc.fault, out.Receipt)
			}
			if tc.wantCode != "" && out.Receipt["reject_code"] != tc.wantCode {
				t.Fatalf("fault %q reject_code = %v, want %s", tc.fault, out.Receipt["reject_code"], tc.wantCode)
			}
			if plant.applyCalls != tc.wantApply {
				t.Fatalf("fault %q plant calls = %d, want %d", tc.fault, plant.applyCalls, tc.wantApply)
			}
			if tc.wantBoot != "" && d.BootID() != tc.wantBoot {
				t.Fatalf("fault %q boot id = %q, want %q", tc.fault, d.BootID(), tc.wantBoot)
			}
			if tc.fault == FaultStuck && d.State()["current_output"].(map[string]any)["energized"] != false {
				t.Fatal("scheduled stuck fault must leave output de-energized")
			}
		})
	}
}

func TestParseFaultSpec(t *testing.T) {
	cases := map[string]FaultInjection{
		"stuck":      {Name: FaultStuck, AcceptedCommand: 1},
		"ack_lost@3": {Name: FaultAckLost, AcceptedCommand: 3},
	}
	for input, want := range cases {
		got, err := ParseFaultSpec(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("parse %q = %+v, want %+v", input, got, want)
		}
	}
	for _, input := range []string{"unknown", "stuck@0", "stuck@x", "stuck@1@2"} {
		if _, err := ParseFaultSpec(input); err == nil {
			t.Fatalf("parse %q should fail", input)
		}
	}
}
