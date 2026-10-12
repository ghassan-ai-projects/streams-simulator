package app

import (
	"maps"
	"slices"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (d *devices) state(device string) []wire.Record {
	id, known := d.entity(device)
	if !known {
		return nil
	}
	kind := d.spec.InstanceKind(d.spec.InstanceSetFor(id))
	values := []any{}
	for _, channel := range slices.Sorted(maps.Keys(d.current[id])) {
		values = append(values, map[string]any{"name": channel, "alias": wire.Alias(kind, channel), "value": d.current[id][channel], "quality": "good"})
	}
	return []wire.Record{wire.DeviceState(device, d.boot(id), values)}
}

func (d *devices) bootOf(device string) (string, bool) {
	id, known := d.entity(device)
	if !known {
		return "none", false
	}
	return d.boot(id), true
}
