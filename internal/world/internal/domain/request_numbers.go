package domain

import "encoding/json"

func numbersByValue(value any) any {
	switch typed := value.(type) {
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return number
		}
	case map[string]any:
		return objectNumbersByValue(typed)
	case []any:
		return listNumbersByValue(typed)
	}
	return value
}

func objectNumbersByValue(object map[string]any) map[string]any {
	out := make(map[string]any, len(object))
	// determinism-safe: builds a map; json.Marshal sorts its keys.
	for key, item := range object {
		out[key] = numbersByValue(item)
	}
	return out
}

func listNumbersByValue(items []any) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = numbersByValue(item)
	}
	return out
}
