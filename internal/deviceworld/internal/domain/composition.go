package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// SelectEntity is the world entity the device targets are bound to: the one
// named, or the first entity of the world when none is.
func SelectEntity(w *world.World, entity string) (string, error) {
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

// ValidateComposition checks the bindings against the world and the device's
// capability catalog before a listener is opened: every bound target and every
// safe stop the catalog declares for a bound target must be executable.
func ValidateComposition(w *world.World, bindings map[string]Binding, caps *device.Capabilities) error {
	targets := slices.Sorted(maps.Keys(bindings))
	stops := safeStops(bindings, caps)
	if err := ValidateBindings(w, bindings, targets, stops); err != nil {
		return fmt.Errorf("validate device world bindings: %w", err)
	}
	return nil
}

func safeStops(bindings map[string]Binding, caps *device.Capabilities) []string {
	stops := make([]string, 0, len(bindings))
	for _, target := range caps.SafeStopNames() {
		if _, ok := bindings[target]; ok {
			stops = append(stops, target)
		}
	}
	return stops
}
