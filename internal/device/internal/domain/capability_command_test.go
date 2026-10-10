package domain

import (
	"os"
	"strings"
	"testing"
)

func TestCapabilityDigestIsBoundToLoadedData(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../testdata/thermal.capabilities.json")
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
// ApplyFrame, and encode the outcome frames. It keeps the wire path under test.
func (d *Device) HandleCommandFor(t *testing.T, command map[string]any) ([]byte, []byte, bool, error) {
	t.Helper()
	frame, err := EncodeRecord(command)
	if err != nil {
		t.Fatalf("encode command: %v", err)
	}
	outcome, err := d.ApplyFrame(frame)
	if err != nil {
		return nil, nil, false, err
	}
	receipt, err := EncodeRecord(outcome.Receipt)
	if err != nil {
		t.Fatalf("encode receipt: %v", err)
	}
	result, err := EncodeRecord(outcome.Result)
	if err != nil {
		t.Fatalf("encode result: %v", err)
	}
	return receipt, result, outcome.AckLost, nil
}

const keyB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
