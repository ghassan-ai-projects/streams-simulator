package deviceworld

import (
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

func setpointBinding(entity string) map[string]Binding {
	return map[string]Binding{
		"fan-01": {
			Effector:   "adjust_setpoint",
			Entity:     entity,
			ValueState: "setpoint_true",
			Args: func(_ map[string]float64) map[string]any {
				return map[string]any{"reefer_id": entity, "setpoint_c": float64(-20)}
			},
		},
	}
}

func plantCommand(commandID string, atMicros int64) device.PlantCommand {
	return device.PlantCommand{
		Target: "fan-01", Operation: "set_pwm_lease",
		Params:    map[string]float64{"duty_permille": 450, "lease_ms": 5000},
		CommandID: commandID, AtMicros: atMicros,
	}
}

// TestEnergizedMirrorsWorldEffect proves the adapter reports the world's physical
// truth, not the command's wish: energized must equal what the world's own
// effector-call ledger recorded as EffectApplied, whatever failure mode fired.
func TestEnergizedMirrorsWorldEffect(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, setpointBinding(entity))
	atMicros := w.Clock() / 1000

	effect := plant.Apply(plantCommand("cmd-1", atMicros))

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
	plant := New(w, setpointBinding(entity))
	atMicros := w.Clock() / 1000

	first := plant.Apply(plantCommand("cmd-dup", atMicros))
	second := plant.Apply(plantCommand("cmd-dup", atMicros))

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
	plant := New(w, setpointBinding(entity))

	effect := plant.Apply(device.PlantCommand{Target: "unknown-99", Operation: "set_pwm_lease", CommandID: "cmd-x", AtMicros: w.Clock() / 1000})

	if effect.Energized {
		t.Fatal("an unmapped target must not energize")
	}
	if len(w.EffectorCalls()) != 0 {
		t.Fatalf("an unmapped target must not invoke the world: %d calls", len(w.EffectorCalls()))
	}
}

// TestWiredThroughDevice proves the full path: a governed device command, routed
// through the emulator's admission logic, drives the world oracle and the
// device's reported state agrees with the world.
func TestWiredThroughDevice(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	plant := New(w, setpointBinding(entity))

	atMicros := w.Clock() / 1000
	d := device.New(device.Config{Plant: plant, Capabilities: deviceCaps(t), Clock: func() int64 { return atMicros }})

	// A valid, in-bounds command bound to the device's boot.
	out := d.ApplyCommand(map[string]any{
		"message_type": "command", "protocol_version": float64(1),
		"command_id": "cmd-through", "idempotency_key": "sha256:" + rep('a', 64),
		"target": "fan-01", "operation": "set_pwm_lease",
		"parameters":       map[string]any{"duty_permille": float64(450), "lease_ms": float64(5000)},
		"expected_boot_id": "boot-A", "not_before_mono_us": float64(atMicros),
		"expires_after_ms": float64(60000), "policy_digest": "sha256:" + rep('b', 64),
	})

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

func rep(b byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return string(out)
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
