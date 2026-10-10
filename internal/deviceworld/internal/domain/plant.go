// Package domain holds the device-to-world binding rules: the data-defined
// binding catalog, its validation against a world spec, and the Plant that
// drives the world's effectors for an accepted device command. It performs no
// I/O.
package domain

import (
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
	binding, err := p.targetBinding(cmd.Target)
	if err != nil {
		return device.PlantEffect{}, err
	}
	atNS, args, err := p.prepareCommand(binding, cmd)
	if err != nil {
		return device.PlantEffect{}, err
	}
	result, err := p.invokeWorld(binding.effector, binding.entity, cmd.CommandID, args, atNS, "world returned no effector result")
	if err != nil {
		return device.PlantEffect{}, err
	}
	return p.completeCommand(binding, result, atNS)
}

// SafeStop invokes the bound world effector with its catalog-owned, literal
// arguments. It is used when the device lease expires or the device reboots;
// clearing the device state alone would leave the world oracle unchanged.
func (p *Plant) SafeStop(target string, atMicros int64) (device.PlantEffect, error) {
	binding, err := p.targetBinding(target)
	if err != nil {
		return device.PlantEffect{}, err
	}
	if binding.safeStopEffector == "" {
		return device.PlantEffect{}, fmt.Errorf("%w: no explicit safe-stop binding for target %q", device.ErrPlantUnavailable, target)
	}
	atNS, args, err := p.prepareSafeStop(binding, atMicros)
	if err != nil {
		return device.PlantEffect{}, err
	}
	return p.invokeSafeStop(binding, target, atNS, args)
}

func (p *Plant) advance(atMicros int64) (int64, error) {
	atNS := p.worldInstant(atMicros)
	if _, _, err := p.w.Advance(atNS); err != nil {
		return 0, fmt.Errorf("advance world to %d: %w", atNS, err)
	}
	return atNS, nil
}

func stateValue(w *world.World, binding Binding) float64 {
	if binding.valueState == "" {
		return 0
	}
	return w.StateValue(binding.entity, binding.valueState, w.Clock())
}

func (p *Plant) worldInstant(atMicros int64) int64 {
	atNS := atMicros * 1000
	// Accept boot-relative device clocks and the historical epoch-relative test form.
	if atNS < p.w.StartNS {
		atNS = p.w.StartNS + atNS
	}
	if atNS < p.w.Clock() {
		atNS = p.w.Clock()
	}
	return atNS
}
