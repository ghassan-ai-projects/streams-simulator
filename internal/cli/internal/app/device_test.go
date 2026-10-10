package app

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestNewDeviceServeDeviceBindsWorldPlant(t *testing.T) {
	t.Parallel()
	dev, worldState, err := newDeviceServeDevice(deviceServeOptions{
		capabilities:  "../../../device/testdata/thermal_capability_catalog.json",
		worldBindings: "../../../deviceworld/testdata/thermal.bindings.json",
		worldDomain:   testsupport.Domain("cold-chain-transit"),
		bootID:        "boot-A", deviceID: "dev-01",
	}, time.Now)
	if err != nil {
		t.Fatalf("build world-backed device: %v", err)
	}
	if worldState == nil {
		t.Fatal("world binding must construct a world")
	}

	commandData, err := os.ReadFile("../../../device/contract/conformance/v1/valid/command.json")
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
	t.Parallel()
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

func TestDeviceWorldClockCountsMicrosecondsOfTheInjectedClock(t *testing.T) {
	t.Parallel()
	now := fixedNow
	clock := func() time.Time { return now }
	_, _, deviceClock, err := prepareDeviceWorld(deviceServeOptions{
		worldBindings: testsupport.RepositoryRoot() + "/internal/deviceworld/testdata/thermal.bindings.json",
		worldDomain:   testsupport.Domain("cold-chain-transit"),
	}, mustCapabilities(t), clock)
	if err != nil {
		t.Fatalf("prepare device world: %v", err)
	}
	if got := deviceClock(); got != 0 {
		t.Fatalf("clock at start = %d, want 0", got)
	}
	now = now.Add(1500 * time.Microsecond)
	if got := deviceClock(); got != 1500 {
		t.Fatalf("clock after 1.5ms = %d, want 1500", got)
	}
}

func mustCapabilities(t *testing.T) *device.Capabilities {
	t.Helper()
	caps, err := loadDeviceCapabilities(testsupport.RepositoryRoot() + "/internal/device/testdata/thermal_capability_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	return caps
}
