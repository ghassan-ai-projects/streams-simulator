package domain

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestLeaseExpiryInvokesWorldSafeStop(t *testing.T) {
	t.Parallel()
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))
	now := w.Clock() / 1000
	d := device.New(device.Config{
		Plant:        plant,
		Capabilities: deviceCaps(t),
		Clock:        func() int64 { return now },
	})

	if out := d.ApplyCommand(deviceCommand(t, "cmd-lease", now)); out.Receipt["accepted"] != true {
		t.Fatalf("valid lease command must be accepted: %v", out.Receipt)
	}
	now += 5_000_001
	state := d.State()
	if state["safe_state"] != true {
		t.Fatalf("expired lease must put device in safe state: %v", state)
	}
	output, ok := state["current_output"].(map[string]any)
	if !ok || output["energized"] != false {
		t.Fatalf("expired lease must de-energize device output: %v", state["current_output"])
	}
	calls := w.EffectorCalls()
	if len(calls) != 2 || calls[1].CommandID != "safe-stop/fan-01" {
		t.Fatalf("lease expiry must invoke the world safe-stop path, calls = %+v", calls)
	}
	if calls[1].Effector != "stop_fan" || w.StateValue(entity, "fan_duty_true", w.Clock()) != 0 {
		t.Fatalf("lease expiry must use the explicit stop_fan effect and clear fan duty: calls=%+v duty=%v", calls, w.StateValue(entity, "fan_duty_true", w.Clock()))
	}
}

func TestWorldSafeStopDoesNotClaimPhysicalEffect(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Domain("cold-chain-transit"))
	if err != nil {
		t.Fatalf("load cold-chain domain: %v", err)
	}
	w, err := world.New(spec, 1, "w-safe-stop-failure", model.DefaultStartTimeNS, world.Options{
		EmitDisabled:     true,
		ForceFailureMode: world.ModeConfirmedNoEffect,
	})
	if err != nil {
		t.Fatalf("new world: %v", err)
	}
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))

	effect, err := plant.SafeStop("fan-01", w.Clock()/1000)
	if !errors.Is(err, device.ErrPlantUnavailable) {
		t.Fatalf("accepted safe stop without a world effect must fail closed: effect=%+v err=%v", effect, err)
	}
	if effect.Energized {
		t.Fatal("failed safe stop must not report an energized effect")
	}
}

func deviceCommand(t *testing.T, commandID string, atMicros int64) map[string]any {
	t.Helper()
	data, err := os.ReadFile("../../../device/contract/conformance/v1/valid/command.json")
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(data, &command); err != nil {
		t.Fatal(err)
	}
	command["command_id"] = commandID
	command["not_before_mono_us"] = float64(atMicros)
	return command
}

// deviceCaps loads the device capability catalog data fixture for the through-
// device test.
func deviceCaps(t *testing.T) *device.Capabilities {
	t.Helper()
	data, err := os.ReadFile("../../../device/testdata/thermal_capability_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	caps, err := device.LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}
