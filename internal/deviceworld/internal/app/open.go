// Package app is the use-case layer of the deviceworld module: it opens a
// device whose plant is a world, in the order the composition must be proven
// (capabilities, fault schedule, bindings, world, bound plant) before any
// listener opens.
package app

import (
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type OpenOptions struct {
	Capabilities, WorldBindings, WorldDomain, WorldEntity, BootID, DeviceID string
	FaultSchedule                                                           []device.FaultInjection
	Now                                                                     func() time.Time
}

func OpenDevice(options OpenOptions) (*device.Device, *world.World, error) {
	caps, err := device.LoadCapabilitiesFile(options.Capabilities)
	if err != nil {
		return nil, nil, fmt.Errorf("%w", err)
	}
	if err := device.ValidateFaultSchedule(options.FaultSchedule); err != nil {
		return nil, nil, fmt.Errorf("validate fault schedule: %w", err)
	}
	plant, worldState, clock, err := preparePlant(options, caps)
	if err != nil {
		return nil, nil, err
	}
	return device.New(deviceConfig(options, caps, plant, clock)), worldState, nil
}

func deviceConfig(options OpenOptions, caps *device.Capabilities, plant device.Plant, clock func() int64) device.Config {
	return device.Config{BootID: options.BootID, DeviceID: options.DeviceID, Capabilities: caps,
		Plant: plant, Clock: clock, FaultSchedule: options.FaultSchedule}
}

func preparePlant(options OpenOptions, caps *device.Capabilities) (device.Plant, *world.World, func() int64, error) {
	if options.WorldBindings == "" {
		return nil, nil, nil, nil
	}
	data, err := files.ReadBindings(options.WorldBindings)
	if err != nil {
		return nil, nil, nil, err
	}
	w, err := createWorld(options.WorldDomain)
	if err != nil {
		return nil, nil, nil, err
	}
	return bindPlant(w, data, options.WorldEntity, caps, options.Now)
}

func createWorld(path string) (*world.World, error) {
	spec, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load device world domain: %w", err)
	}
	w, err := world.New(spec, 1, "device-world", model.DefaultStartTimeNS, world.Options{EmitDisabled: true})
	if err != nil {
		return nil, fmt.Errorf("create device world: %w", err)
	}
	return w, nil
}

func bindPlant(w *world.World, data []byte, entity string, caps *device.Capabilities, now func() time.Time) (device.Plant, *world.World, func() int64, error) {
	entity, err := rules.SelectEntity(w, entity)
	if err != nil {
		return nil, nil, nil, err
	}
	bindings, err := rules.LoadBindings(data, entity)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load device world bindings: %w", err)
	}
	if err := rules.ValidateComposition(w, bindings, caps); err != nil {
		return nil, nil, nil, err
	}
	return rules.New(w, bindings), w, deviceClock(now), nil
}

func deviceClock(now func() time.Time) func() int64 {
	started := now()
	return func() int64 { return now().Sub(started).Microseconds() }
}
