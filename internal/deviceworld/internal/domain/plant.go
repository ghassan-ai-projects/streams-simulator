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

type Binding struct {
	effector         string
	entity           string
	valueState       string
	arguments        map[string]bindingArgument
	safeStopEffector string
	safeStopArgs     map[string]bindingArgument
}

type Plant struct {
	w        *world.World
	bindings map[string]Binding

	safeStops map[string]int
}

func New(w *world.World, bindings map[string]Binding) *Plant {
	return &Plant{w: w, bindings: bindings, safeStops: map[string]int{}}
}

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

	if atNS < p.w.StartNS {
		atNS = p.w.StartNS + atNS
	}
	if atNS < p.w.Clock() {
		atNS = p.w.Clock()
	}
	return atNS
}
