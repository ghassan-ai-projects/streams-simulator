package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

func ScrubValue[T any](value T, roots []string) (T, error) {
	var out T
	tree, err := jsonTree(value)
	if err != nil {
		return out, err
	}
	raw, err := json.Marshal(scrubTree(tree, roots))
	if err != nil {
		return out, fmt.Errorf("canvasbridge: scrub: %w", err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("canvasbridge: scrub: %w", err)
	}
	return out, nil
}

func jsonTree(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: scrub: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var tree any
	if err := decoder.Decode(&tree); err != nil {
		return nil, fmt.Errorf("canvasbridge: scrub: %w", err)
	}
	return tree, nil
}

func scrubTree(node any, roots []string) any {
	switch value := node.(type) {
	case string:
		return Scrub(value, roots)
	case []any:
		for i := range value {
			value[i] = scrubTree(value[i], roots)
		}
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(value)) {
			value[key] = scrubTree(value[key], roots)
		}
	}
	return node
}
