package device

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	physicalCapabilityDigest = "sha256:0d61225286c628cfba8cbf7aea514e1fdc95918b514b4b810516dbe0fc44fc76"
	physicalFirmwareDigest   = "sha256:36a91963b13e5b86749953f8e5441c63c22866ae71796d889fbc629d48d176af"
)

func TestPhysicalArduinoCatalogExecutesMaterializedAlertLEDAndFan(t *testing.T) {
	root := os.Getenv("REAL_WORLD_SENSOR_ROOT")
	if root == "" {
		root = filepath.Join("..", "..", "..", "agent-research-lab", "real-world-sensor")
	}
	// REAL_WORLD_SENSOR_ROOT is an explicit local integration-fixture root; the
	// catalog path below has a fixed suffix and is not request-derived input.
	data, err := os.ReadFile(filepath.Join(root, "assessment", "arduino-mega-l293d-fan-led-capability-catalog.json")) // #nosec G703 -- explicit local cross-repository fixture root
	if err != nil {
		t.Fatalf("read physical catalog: %v", err)
	}
	caps, err := LoadCapabilities(data)
	if err != nil {
		t.Fatalf("load physical catalog: %v", err)
	}
	if got := caps.Digest(); got != physicalCapabilityDigest {
		t.Fatalf("physical catalog digest = %s, want %s", got, physicalCapabilityDigest)
	}

	d := New(Config{
		DeviceID:         "arduino-mega-01",
		BootID:           "boot-cross",
		FirmwareDigest:   physicalFirmwareDigest,
		CapabilityDigest: physicalCapabilityDigest,
		Capabilities:     caps,
		Clock:            func() int64 { return 0 },
	})
	outcome := d.ApplyCommand(map[string]any{
		"message_type":       "command",
		"protocol_version":   float64(1),
		"command_id":         "cross-repo-led",
		"idempotency_key":    "sha256:" + strings.Repeat("a", 64),
		"target":             "led-01",
		"operation":          "set_led",
		"parameters":         map[string]any{"brightness_permille": float64(1000), "pattern": "solid"},
		"expected_boot_id":   "boot-cross",
		"not_before_mono_us": float64(0),
		"expires_after_ms":   float64(60000),
		"policy_digest":      "sha256:" + strings.Repeat("b", 64),
	})
	if outcome.Receipt["accepted"] != true || outcome.Result["status"] != "executed" {
		t.Fatalf("physical LED command outcome = receipt %v result %v", outcome.Receipt, outcome.Result)
	}
	state := d.State()
	if state["firmware_digest"] != physicalFirmwareDigest || state["capability_digest"] != physicalCapabilityDigest {
		t.Fatalf("physical identity = %v/%v", state["firmware_digest"], state["capability_digest"])
	}
	output, ok := state["current_output"].(map[string]any)
	if !ok || output["target"] != "led-01" || output["operation"] != "set_led" || output["value"] != float64(1000) || output["energized"] != true {
		t.Fatalf("physical LED observed output = %v", state["current_output"])
	}

	safeStop := d.ApplyCommand(map[string]any{
		"message_type":       "command",
		"protocol_version":   float64(1),
		"command_id":         "safe-stop/led-01",
		"idempotency_key":    "sha256:" + strings.Repeat("c", 64),
		"target":             "led-01",
		"operation":          "safe_stop",
		"parameters":         map[string]any{},
		"expected_boot_id":   "boot-cross",
		"not_before_mono_us": float64(0),
		"expires_after_ms":   float64(1000),
		"policy_digest":      physicalCapabilityDigest,
	})
	if safeStop.Receipt["accepted"] != true || safeStop.Result["status"] != "safe_state" {
		t.Fatalf("physical safe stop outcome = receipt %v result %v", safeStop.Receipt, safeStop.Result)
	}
	state = d.State()
	output, _ = state["current_output"].(map[string]any)
	if state["safe_state"] != true || output["energized"] != false {
		t.Fatalf("physical safe-stop state = %v", state)
	}

	fan := d.ApplyCommand(map[string]any{
		"message_type":       "command",
		"protocol_version":   float64(1),
		"command_id":         "cross-repo-fan",
		"idempotency_key":    "sha256:" + strings.Repeat("d", 64),
		"target":             "fan-01",
		"operation":          "set_pwm_lease",
		"parameters":         map[string]any{"duty_permille": float64(450), "lease_ms": float64(5000)},
		"expected_boot_id":   "boot-cross",
		"not_before_mono_us": float64(0),
		"expires_after_ms":   float64(20000),
		"policy_digest":      "sha256:" + strings.Repeat("e", 64),
	})
	if fan.Receipt["accepted"] != true || fan.Result["status"] != "executed" {
		t.Fatalf("physical fan command outcome = receipt %v result %v", fan.Receipt, fan.Result)
	}
	state = d.State()
	output, ok = state["current_output"].(map[string]any)
	if !ok || output["target"] != "fan-01" || output["operation"] != "set_pwm_lease" || output["value"] != float64(450) || output["energized"] != true {
		t.Fatalf("physical fan observed output = %v", state["current_output"])
	}

	fanSafeStop := d.ApplyCommand(map[string]any{
		"message_type":       "command",
		"protocol_version":   float64(1),
		"command_id":         "safe-stop/fan-01",
		"idempotency_key":    "sha256:" + strings.Repeat("f", 64),
		"target":             "fan-01",
		"operation":          "safe_stop",
		"parameters":         map[string]any{},
		"expected_boot_id":   "boot-cross",
		"not_before_mono_us": float64(0),
		"expires_after_ms":   float64(1000),
		"policy_digest":      physicalCapabilityDigest,
	})
	if fanSafeStop.Receipt["accepted"] != true || fanSafeStop.Result["status"] != "safe_state" {
		t.Fatalf("physical fan safe stop outcome = receipt %v result %v", fanSafeStop.Receipt, fanSafeStop.Result)
	}
	state = d.State()
	output, _ = state["current_output"].(map[string]any)
	if state["safe_state"] != true || output["target"] != "fan-01" || output["energized"] != false {
		t.Fatalf("physical fan safe-stop state = %v", state)
	}

	// Keep the source catalog parseable as the exact JSON artifact used by both
	// Agentic Stream and this simulator test.
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("physical catalog JSON: %v", err)
	}
}
