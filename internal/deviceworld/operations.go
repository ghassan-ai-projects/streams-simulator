package deviceworld

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func OpenDevice(options OpenOptions) (*device.Device, *world.World, error) {
	return app.OpenDevice(options)
}

func LoadBindings(data []byte, entity string) (map[string]Binding, error) {
	return layer.LoadBindings(data, entity)
}

func ValidateBindings(w *world.World, bindings map[string]Binding, requiredTargets, requiredSafeStops []string) error {
	return layer.ValidateBindings(w, bindings, requiredTargets, requiredSafeStops)
}

func (p *Plant) Apply(cmd device.PlantCommand) (device.PlantEffect, error) {
	if p == nil || p.plant == nil {
		return device.PlantEffect{}, ErrNoPlant
	}
	return p.plant.Apply(cmd)
}

func (p *Plant) SafeStop(target string, atMicros int64) (device.PlantEffect, error) {
	if p == nil || p.plant == nil {
		return device.PlantEffect{}, ErrNoPlant
	}
	return p.plant.SafeStop(target, atMicros)
}
