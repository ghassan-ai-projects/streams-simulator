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
	effector         string
	entity           string
	valueState       string
	arguments        map[string]bindingArgument
	safeStopEffector string
	safeStopArgs     map[string]bindingArgument
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
// device layer handles separately. An unmapped target is unavailable, not a
// successful no-op, so the device cannot report execution without an effect.
func (p *Plant) Apply(cmd device.PlantCommand) (device.PlantEffect, error) {
	binding, ok := p.bindings[cmd.Target]
	if !ok {
		return device.PlantEffect{}, fmt.Errorf("%w: no binding for target %q", device.ErrPlantUnavailable, cmd.Target)
	}
	if p.w == nil {
		return device.PlantEffect{}, fmt.Errorf("%w: world is unavailable", device.ErrPlantUnavailable)
	}
	atNS, err := p.advance(cmd.AtMicros)
	if err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world: %w", device.ErrPlantUnavailable, err)
	}
	args, err := binding.args(cmd.Params)
	if err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: build effector arguments: %w", device.ErrPlantUnavailable, err)
	}
	res, err := p.w.InvokeEffector(binding.effector, binding.entity, cmd.CommandID, args, atNS)
	if err != nil || res == nil {
		if errors.Is(err, world.ErrInterlockRefused) {
			return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantInterlocked, err)
		}
		if err == nil {
			err = fmt.Errorf("world returned no effector result")
		}
		return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantUnavailable, err)
	}
	if _, _, err := p.w.Advance(atNS + 1); err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world after effector: %w", device.ErrPlantUnavailable, err)
	}
	effect := device.PlantEffect{Energized: res.EffectApplied}
	if binding.valueState != "" {
		effect.Value = p.w.StateValue(binding.entity, binding.valueState, p.w.Clock())
	}
	return effect, nil
}

// SafeStop invokes the bound world effector with its catalog-owned, literal
// arguments. It is used when the device lease expires or the device reboots;
// clearing the device state alone would leave the world oracle unchanged.
func (p *Plant) SafeStop(target string, atMicros int64) (device.PlantEffect, error) {
	binding, ok := p.bindings[target]
	if !ok {
		return device.PlantEffect{}, fmt.Errorf("%w: no binding for target %q", device.ErrPlantUnavailable, target)
	}
	if binding.safeStopEffector == "" {
		return device.PlantEffect{}, fmt.Errorf("%w: no explicit safe-stop binding for target %q", device.ErrPlantUnavailable, target)
	}
	if p.w == nil {
		return device.PlantEffect{}, fmt.Errorf("%w: world is unavailable", device.ErrPlantUnavailable)
	}
	atNS, err := p.advance(atMicros)
	if err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world for safe stop: %w", device.ErrPlantUnavailable, err)
	}
	args, err := binding.safeStopArgsForEntity()
	if err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: build safe-stop arguments: %w", device.ErrPlantUnavailable, err)
	}
	res, err := p.w.InvokeEffector(binding.safeStopEffector, binding.entity, "safe-stop/"+target, args, atNS)
	if err != nil || res == nil {
		if errors.Is(err, world.ErrInterlockRefused) {
			return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantInterlocked, err)
		}
		if err == nil {
			err = fmt.Errorf("world returned no safe-stop result")
		}
		return device.PlantEffect{}, fmt.Errorf("%w: %w", device.ErrPlantUnavailable, err)
	}
	if !res.Accepted {
		return device.PlantEffect{}, fmt.Errorf("%w: safe stop was not accepted: %s", device.ErrPlantUnavailable, res.Reason)
	}
	if !res.EffectApplied {
		return device.PlantEffect{}, fmt.Errorf("%w: safe stop was accepted without applying its effect: %s", device.ErrPlantUnavailable, res.Mode)
	}
	if _, _, err := p.w.Advance(atNS + 1); err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world after safe stop: %w", device.ErrPlantUnavailable, err)
	}
	return device.PlantEffect{Value: stateValue(p.w, binding), Energized: false}, nil
}

func (p *Plant) advance(atMicros int64) (int64, error) {
	atNS := atMicros * 1000
	// Device clocks are boot-relative. Focused plant tests historically pass
	// epoch-relative microseconds, so retain that form while accepting the
	// relative clock used by the CLI gateway.
	if atNS < p.w.StartNS {
		atNS = p.w.StartNS + atNS
	}
	if atNS < p.w.Clock() {
		atNS = p.w.Clock()
	}
	if _, _, err := p.w.Advance(atNS); err != nil {
		return 0, err
	}
	return atNS, nil
}

func stateValue(w *world.World, binding Binding) float64 {
	if binding.valueState == "" {
		return 0
	}
	return w.StateValue(binding.entity, binding.valueState, w.Clock())
}
