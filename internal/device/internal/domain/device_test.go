package domain

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
	data, err := os.ReadFile("../../testdata/thermal_capability_catalog.json")
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
	raw, err := os.ReadFile("../../contract/conformance/v1/valid/command.json")
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
	t.Parallel()
	d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return 0 }})
	out := d.ApplyCommand(validCommand(t, func(c map[string]any) {
		c["not_before_mono_us"] = float64(1)
	}))
	if out.Receipt["accepted"] != false || out.Receipt["reject_code"] != "not_ready" {
		t.Fatalf("future-bound command must be rejected as not_ready: %v", out.Receipt)
	}
}

func TestLeaseExpiryReturnsSafeState(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	for _, reboot := range []bool{false, true} {
		name := "lease expiry"
		if reboot {
			name = "reboot"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
