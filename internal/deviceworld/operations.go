package deviceworld

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// OpenDevice builds the device the options describe, proving the composition
// before any listener opens. With WorldBindings the device's plant is the world
// bound by that catalog, which is returned too; without, the device has no
// plant and the world is nil.
func OpenDevice(options OpenOptions) (*device.Device, *world.World, error) {
	return app.OpenDevice(options)
}

// LoadBindings parses a strict device-target to world-effector binding
// catalog. Entity-source arguments resolve to the supplied runtime world
// entity; all other mapping and literal values come from the JSON catalog.
func LoadBindings(data []byte, entity string) (map[string]Binding, error) {
	return layer.LoadBindings(data, entity)
}

// ValidateBindings checks a complete device/world composition before a
// listener is opened: its names and argument shapes agree with the world.
func ValidateBindings(w *world.World, bindings map[string]Binding, requiredTargets, requiredSafeStops []string) error {
	return layer.ValidateBindings(w, bindings, requiredTargets, requiredSafeStops)
}

// Apply invokes the bound world effector for an accepted device command and
// reports the physical truth. An unmapped target is unavailable, not a
// successful no-op.
func (p *Plant) Apply(cmd device.PlantCommand) (device.PlantEffect, error) {
	if p == nil || p.plant == nil {
		return device.PlantEffect{}, ErrNoPlant
	}
	return p.plant.Apply(cmd)
}

// SafeStop drives the target to its declared safe state at atMicros.
func (p *Plant) SafeStop(target string, atMicros int64) (device.PlantEffect, error) {
	if p == nil || p.plant == nil {
		return device.PlantEffect{}, ErrNoPlant
	}
	return p.plant.SafeStop(target, atMicros)
}
