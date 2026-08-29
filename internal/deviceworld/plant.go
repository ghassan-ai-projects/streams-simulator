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
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Binding maps one device target to a world effector invocation.
type Binding struct {
	// Effector is the world effector name to invoke.
	Effector string
	// Entity is the world entity id the effector acts on.
	Entity string
	// ValueState, when set, is the world state read back as the output's value
	// — the independent process reading used for verification.
	ValueState string
	// Args builds the world effector arguments from the device's numeric
	// parameters. It must satisfy the effector's args_schema.
	Args func(params map[string]float64) map[string]any
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
func (p *Plant) Apply(cmd device.PlantCommand) device.PlantEffect {
	binding, ok := p.bindings[cmd.Target]
	if !ok {
		return device.PlantEffect{}
	}
	args := map[string]any{}
	if binding.Args != nil {
		args = binding.Args(cmd.Params)
	}
	res, err := p.w.InvokeEffector(binding.Effector, binding.Entity, cmd.CommandID, args, cmd.AtMicros*1000)
	if err != nil || res == nil {
		// Interlock refusal, unknown entity/effector, or invalid args: the
		// world applied nothing, so the output is not energized.
		return device.PlantEffect{}
	}
	effect := device.PlantEffect{Energized: res.EffectApplied}
	if binding.ValueState != "" {
		effect.Value = p.w.StateValue(binding.Entity, binding.ValueState, p.w.Clock())
	}
	return effect
}
