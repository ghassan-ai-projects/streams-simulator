package deviceworld

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// coldChainWorld builds a world from the shipped cold-chain domain, whose
// adjust_setpoint effector reduces the setpoint_true state — a real effector
// oracle to drive the device plant against.
func coldChainWorld(t *testing.T, seed uint64) *world.World {
	t.Helper()
	spec, err := domain.Load("../../domains/cold-chain-transit.domain.json")
	if err != nil {
		t.Fatalf("load cold-chain domain: %v", err)
	}
	w, err := world.New(spec, seed, "w-deviceworld", model.DefaultStartTimeNS, world.Options{EmitDisabled: true})
	if err != nil {
		t.Fatalf("new world: %v", err)
	}
	return w
}

func loadBindings(t *testing.T, entity string) map[string]Binding {
	t.Helper()
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(data, entity)
	if err != nil {
		t.Fatal(err)
	}
	return bindings
}

func plantCommand(t *testing.T, commandID string, atMicros int64) device.PlantCommand {
	t.Helper()
	command := deviceCommand(t, commandID, atMicros)
	parameters, ok := command["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("command parameters have wrong type: %T", command["parameters"])
	}
	numeric := make(map[string]float64, len(parameters))
	for name, raw := range parameters {
		value, ok := raw.(float64)
		if !ok {
			t.Fatalf("command parameter %q has wrong type: %T", name, raw)
		}
		numeric[name] = value
	}
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	return device.PlantCommand{
		Target: target, Operation: operation, Params: numeric,
		CommandID: commandID, AtMicros: atMicros,
	}
}

// TestEnergizedMirrorsWorldEffect proves the adapter reports the world's physical
// truth, not the command's wish: energized must equal what the world's own
// effector-call ledger recorded as EffectApplied, whatever failure mode fired.
func TestEnergizedMirrorsWorldEffect(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))
	atMicros := w.Clock() / 1000

	effect, err := plant.Apply(plantCommand(t, "cmd-1", atMicros))
	if err != nil {
		t.Fatalf("apply plant command: %v", err)
	}

	calls := w.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly one effector call, got %d", len(calls))
	}
	if effect.Energized != calls[0].EffectApplied {
		t.Fatalf("energized (%v) must mirror the world's EffectApplied (%v)", effect.Energized, calls[0].EffectApplied)
	}
}

// TestIdempotentByCommandID proves a replayed device command_id does not apply a
// second world effect — the ledger stays at one call and the reported effect is
// stable.
func TestIdempotentByCommandID(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))
	atMicros := w.Clock() / 1000

	first, err := plant.Apply(plantCommand(t, "cmd-dup", atMicros))
	if err != nil {
		t.Fatalf("apply first plant command: %v", err)
	}
	second, err := plant.Apply(plantCommand(t, "cmd-dup", atMicros))
	if err != nil {
		t.Fatalf("apply duplicate plant command: %v", err)
	}

	if len(w.EffectorCalls()) != 1 {
		t.Fatalf("a replayed command_id must not add a second effector call: %d", len(w.EffectorCalls()))
	}
	if first.Energized != second.Energized {
		t.Fatal("replayed command must report the same effect")
	}
}

// TestUnmappedTargetFailsSafe proves an unbound target energizes nothing and
// touches the world not at all.
func TestUnmappedTargetFailsSafe(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))

	command := plantCommand(t, "cmd-x", w.Clock()/1000)
	command.Target = "unknown-99"
	effect, err := plant.Apply(command)
	if err != nil {
		t.Fatalf("unmapped target should be a no-op: %v", err)
	}

	if effect.Energized {
		t.Fatal("an unmapped target must not energize")
	}
	if len(w.EffectorCalls()) != 0 {
		t.Fatalf("an unmapped target must not invoke the world: %d calls", len(w.EffectorCalls()))
	}
}

func TestStuckDeviceDoesNotApplyWorldEffect(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))
	atMicros := w.Clock() / 1000
	d := device.New(device.Config{
		Plant: plant, Capabilities: deviceCaps(t),
		Clock: func() int64 { return atMicros },
	})
	d.SetFaults(device.Faults{Stuck: true})

	out := d.ApplyCommand(deviceCommand(t, "cmd-stuck", atMicros))
	if out.Receipt["accepted"] != true {
		t.Fatalf("stuck command should be acknowledged: %v", out.Receipt)
	}
	if len(w.EffectorCalls()) != 0 {
		t.Fatalf("stuck device must not mutate the world: %d calls", len(w.EffectorCalls()))
	}
}

func TestWorldBindingFailureIsNotReportedAsExecution(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	bindings := loadBindings(t, entity)
	command := deviceCommand(t, "cmd-unavailable", w.Clock()/1000)
	target, _ := command["target"].(string)
	binding := bindings[target]
	binding.effector = "missing-effector"
	bindings[target] = binding
	atMicros := w.Clock() / 1000
	d := device.New(device.Config{
		Plant: New(w, bindings), Capabilities: deviceCaps(t),
		Clock: func() int64 { return atMicros },
	})

	command["not_before_mono_us"] = float64(atMicros)
	out := d.ApplyCommand(command)
	if out.Receipt["accepted"] != false || out.Receipt["reject_code"] != "not_ready" {
		t.Fatalf("binding failure must be a terminal not_ready rejection: %v", out.Receipt)
	}
	if out.Result["status"] != "rejected" {
		t.Fatalf("binding failure must not be reported as executed: %v", out.Result)
	}
}

// TestWiredThroughDevice proves the full path: a governed device command, routed
// through the emulator's admission logic, drives the world oracle and the
// device's reported state agrees with the world.
func TestWiredThroughDevice(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, loadBindings(t, entity))

	atMicros := w.Clock() / 1000
	d := device.New(device.Config{Plant: plant, Capabilities: deviceCaps(t), Clock: func() int64 { return atMicros }})

	// A valid, in-bounds command bound to the device's boot.
	out := d.ApplyCommand(deviceCommand(t, "cmd-through", atMicros))

	if out.Receipt["accepted"] != true {
		t.Fatalf("valid command must be accepted: %v", out.Receipt)
	}
	calls := w.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("device command must drive exactly one world effector call, got %d", len(calls))
	}
	output, _ := d.State()["current_output"].(map[string]any)
	if output == nil || output["energized"] != calls[0].EffectApplied {
		t.Fatalf("device current_output.energized must agree with the world oracle: %v vs %v", output, calls[0].EffectApplied)
	}
}

func deviceCommand(t *testing.T, commandID string, atMicros int64) map[string]any {
	t.Helper()
	data, err := os.ReadFile("../device/contract/conformance/v1/valid/command.json")
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
	data, err := os.ReadFile("../device/testdata/thermal.capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	caps, err := device.LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}
