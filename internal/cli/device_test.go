package cli

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNewDeviceServeDeviceBindsWorldPlant(t *testing.T) {
	dev, worldState, err := newDeviceServeDevice(
		"../../internal/device/testdata/thermal_capability_catalog.json",
		"../../internal/deviceworld/testdata/thermal.bindings.json",
		"../../domains/cold-chain-transit.domain.json",
		"", "boot-A", "dev-01", nil,
	)
	if err != nil {
		t.Fatalf("build world-backed device: %v", err)
	}
	if worldState == nil {
		t.Fatal("world binding must construct a world")
	}

	commandData, err := os.ReadFile("../../internal/device/contract/conformance/v1/valid/command.json")
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(commandData, &command); err != nil {
		t.Fatal(err)
	}
	command["not_before_mono_us"] = float64(0)
	out := dev.ApplyCommand(command)
	if out.Receipt["accepted"] != true {
		t.Fatalf("valid command must be accepted: %v", out.Receipt)
	}
	calls := worldState.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("device command must invoke world effector once, got %d", len(calls))
	}
	output, ok := dev.State()["current_output"].(map[string]any)
	if !ok || output["energized"] != calls[0].EffectApplied {
		t.Fatalf("device state must report world truth: output=%v calls=%+v", output, calls)
	}
}

func TestFaultSpecFlagIsRepeatable(t *testing.T) {
	var flag faultSpecFlag
	if err := flag.Set("stuck@2"); err != nil {
		t.Fatal(err)
	}
	if err := flag.Set("ack_lost"); err != nil {
		t.Fatal(err)
	}
	if len(flag.entries) != 2 || flag.entries[0].AcceptedCommand != 2 || flag.entries[1].AcceptedCommand != 1 {
		t.Fatalf("fault entries = %+v", flag.entries)
	}
}
