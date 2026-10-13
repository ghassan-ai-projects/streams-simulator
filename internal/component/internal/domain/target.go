package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

var keyedItems = map[string]string{"inputs": "state", "affects": "state", "state_deltas": "state", "failure_modes": "mode"}

var sectionKeys = map[string]string{"state": "name", "dynamics": "target", "flows": "name", "effectors": "name", "faults": "id", "measurement_points": "name"}

type workspace struct {
	kind   map[string]any
	points []any
}

func newWorkspace(equipment *Equipment) *workspace {
	return &workspace{kind: cloneMap(equipment.Document["kind"]), points: cloneList(equipment.Document["measurement_points"])}
}

type slot struct {
	object map[string]any
	list   []any
	key    string
	index  int
}

func (s slot) get() any {
	if s.object != nil {
		return s.object[s.key]
	}
	return s.list[s.index]
}

func (s slot) set(value any) {
	if s.object != nil {
		s.object[s.key] = value
		return
	}
	s.list[s.index] = value
}

func (w *workspace) locate(target string) (slot, error) {
	parts := strings.Split(target, "/")
	if len(parts) < 3 {
		return slot{}, fmt.Errorf("target %q is not <section>/<name>/<path>", target)
	}
	entry, err := w.entry(parts[0], parts[1])
	if err != nil {
		return slot{}, fmt.Errorf("target %q: %w", target, err)
	}
	found, err := walk(entry, parts[2:])
	if err != nil {
		return slot{}, fmt.Errorf("target %q: %w", target, err)
	}
	return found, nil
}

func (w *workspace) entry(section, name string) (map[string]any, error) {
	var items []any
	if section == "measurement_points" {
		items = w.points
	} else {
		items, _ = w.kind[section].([]any)
	}
	for _, item := range items {
		if entry, ok := item.(map[string]any); ok && entry[sectionKeys[section]] == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("%s has no entry %q", section, name)
}

func walk(entry map[string]any, path []string) (slot, error) {
	var current any = entry
	var parent string
	var found slot
	for _, segment := range path {
		var err error
		if found, err = step(current, parent, segment); err != nil {
			return slot{}, err
		}
		current, parent = found.get(), found.name(segment)
	}
	return found, scalarLeaf(found.get(), path[len(path)-1])
}

func (s slot) name(segment string) string {
	if s.object == nil {
		return ""
	}
	return segment
}

func step(current any, parent, segment string) (slot, error) {
	switch typed := current.(type) {
	case map[string]any:
		if _, ok := typed[segment]; !ok {
			return slot{}, fmt.Errorf("no member %q", segment)
		}
		return slot{object: typed, key: segment}, nil
	case []any:
		index, err := itemIndex(typed, parent, segment)
		return slot{list: typed, index: index}, err
	}
	return slot{}, fmt.Errorf("%q is below a value that has no members", segment)
}

func itemIndex(items []any, parent, segment string) (int, error) {
	key, keyed := keyedItems[parent]
	if !keyed {
		return position(items, segment)
	}
	for index, item := range items {
		if entry, ok := item.(map[string]any); ok && entry[key] == segment {
			return index, nil
		}
	}
	return 0, fmt.Errorf("no item with %s %q", key, segment)
}

func position(items []any, segment string) (int, error) {
	index, err := strconv.Atoi(segment)
	if err != nil || index < 0 || index >= len(items) {
		return 0, fmt.Errorf("position %q is outside the %d items", segment, len(items))
	}
	return index, nil
}

func scalarLeaf(value any, name string) error {
	switch value.(type) {
	case map[string]any, []any:
		return fmt.Errorf("%q is an object or list, not a single value", name)
	}
	return nil
}

func (w *workspace) write(target string, value any) error {
	found, err := w.locate(target)
	if err != nil {
		return err
	}
	if kindOfValue(found.get()) != kindOfValue(value) {
		return fmt.Errorf("target %q holds a %s, not a %s", target, kindOfValue(found.get()), kindOfValue(value))
	}
	found.set(value)
	return nil
}

func kindOfValue(value any) string {
	switch value.(type) {
	case json.Number, float64, int:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	}
	return "value"
}

func cloneMap(value any) map[string]any {
	source, _ := value.(map[string]any)
	out := make(map[string]any, len(source))
	for _, key := range slices.Sorted(maps.Keys(source)) {
		out[key] = cloneValue(source[key])
	}
	return out
}

func cloneList(value any) []any {
	source, _ := value.([]any)
	out := make([]any, len(source))
	for i, item := range source {
		out[i] = cloneValue(item)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		return cloneList(typed)
	}
	return value
}
