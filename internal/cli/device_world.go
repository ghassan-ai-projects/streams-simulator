package cli

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
	"os"
	"time"
)

func loadDeviceCapabilities(path string) (*device.Capabilities, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read capabilities: %w", err)
	}
	caps, err := device.LoadCapabilities(data)
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	return caps, nil
}

func createDeviceWorld(path string) (*world.World, error) {
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

func bindDeviceWorld(w *world.World, data []byte, entity string, caps *device.Capabilities) (device.Plant, *world.World, func() int64, error) {
	entity, err := selectDeviceWorldEntity(w, entity)
	if err != nil {
		return nil, nil, nil, err
	}
	bindings, err := deviceworld.LoadBindings(data, entity)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load device world bindings: %w", err)
	}
	if err := validateDeviceWorldBindings(w, bindings, caps); err != nil {
		return nil, nil, nil, err
	}
	return deviceWorldPlant(w, bindings)
}

func selectDeviceWorldEntity(w *world.World, entity string) (string, error) {
	if entity == "" {
		ids := w.EntityIDs()
		if len(ids) == 0 {
			return "", fmt.Errorf("device world has no entities")
		}
		entity = ids[0]
	}
	if w.Entity(entity) == nil {
		return "", fmt.Errorf("device world entity %q does not exist", entity)
	}
	return entity, nil
}

func validateDeviceWorldBindings(w *world.World, bindings map[string]deviceworld.Binding, caps *device.Capabilities) error {
	targets := make([]string, 0, len(bindings))
	for target := range bindings {
		targets = append(targets, target)
	}
	stops := deviceWorldSafeStops(bindings, caps)
	if err := deviceworld.ValidateBindings(w, bindings, targets, stops); err != nil {
		return fmt.Errorf("validate device world bindings: %w", err)
	}
	return nil
}

func deviceWorldPlant(w *world.World, bindings map[string]deviceworld.Binding) (device.Plant, *world.World, func() int64, error) {
	plant, err := deviceworld.New(w, bindings)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("device world plant: %w", err)
	}
	started := time.Now()
	clock := func() int64 { return time.Since(started).Microseconds() }
	return plant, w, clock, nil
}

func deviceWorldSafeStops(bindings map[string]deviceworld.Binding, caps *device.Capabilities) []string {
	stops := make([]string, 0, len(bindings))
	for _, target := range caps.SafeStopNames() {
		if _, ok := bindings[target]; ok {
			stops = append(stops, target)
		}
	}
	return stops
}
