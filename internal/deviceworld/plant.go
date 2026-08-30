// Package deviceworld binds the device emulator's Plant seam to the Streams
// Simulator world, so the world is the effector oracle behind the wire loop.
//
// With this plant, a device command does not "energize" an output because a
// heuristic says duty>0; it energizes because the world's effector actually
// applied a physical effect. The world's failure modes then become the device's
// observed truth: a confirmed-no-effect or interlock refusal surfaces as
// energized=false — the desired≠observed case — while a silent-no-effect fools
// the device's self-reported state exactly as it would fool a real confirmation
// channel, and only the independent process reading (Value) betrays it.
package deviceworld

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Binding is a validated data-defined mapping from one device target to a world
// effector invocation. LoadBindings is the supported construction path; domain
// mappings are not supplied as Go callbacks.
type Binding struct {
	effector   string
	entity     string
	valueState string
	arguments  map[string]bindingArgument
}

// Plant is a device.Plant backed by a *world.World.
type Plant struct {
	w        *world.World
	bindings map[string]Binding
}

// New returns a world-backed plant. bindings maps device targets (e.g.
// "fan-01") to world effector invocations.
func New(w *world.World, bindings map[string]Binding) *Plant {
	return &Plant{w: w, bindings: bindings}
}

// Apply invokes the bound world effector for an accepted device command and
// reports the physical truth. energized reflects whether the world applied the
// effect (res.EffectApplied) — independent of the acknowledgement, which the
// device layer handles separately. An unmapped target is a fail-safe no-op.
func (p *Plant) Apply(cmd device.PlantCommand) (device.PlantEffect, error) {
	binding, ok := p.bindings[cmd.Target]
	if !ok {
		return device.PlantEffect{}, nil
	}
	if p.w == nil {
		return device.PlantEffect{}, fmt.Errorf("%w: world is unavailable", device.ErrPlantUnavailable)
	}
	args, err := binding.args(cmd.Params)
	if err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: build effector arguments: %w", device.ErrPlantUnavailable, err)
	}
	res, err := p.w.InvokeEffector(binding.effector, binding.entity, cmd.CommandID, args, cmd.AtMicros*1000)
	if err != nil || res == nil {
		if errors.Is(err, world.ErrInterlockRefused) {
			return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantInterlocked, err)
		}
		if err == nil {
			err = fmt.Errorf("world returned no effector result")
		}
		return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantUnavailable, err)
	}
	effect := device.PlantEffect{Energized: res.EffectApplied}
	if binding.valueState != "" {
		effect.Value = p.w.StateValue(binding.entity, binding.valueState, p.w.Clock())
	}
	return effect, nil
}
