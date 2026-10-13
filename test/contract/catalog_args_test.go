package contract

import (
	"encoding/json"
	"math"
	"slices"
)

const catalogArgAttempts = 4

// catalogArgs builds arguments that satisfy an effector's schema; each attempt picks another end of every range.
func catalogArgs(schema map[string]any, entity string, attempt int) map[string]any {
	args := map[string]any{}
	props, _ := schema["properties"].(map[string]any)
	for _, name := range sortedKeys(props) {
		prop, _ := props[name].(map[string]any)
		args[name] = catalogArg(name, prop, entity, attempt)
	}
	return args
}

func catalogArg(name string, prop map[string]any, entity string, attempt int) any {
	if values, ok := prop["enum"].([]any); ok && len(values) > 0 {
		return values[attempt%len(values)]
	}
	switch prop["type"] {
	case "string":
		if len(name) > 3 && name[len(name)-3:] == "_id" {
			return entity
		}
		return "x"
	case "boolean":
		return attempt%2 == 0
	case "integer":
		return math.Round(numericArg(prop, attempt))
	}
	return numericArg(prop, attempt)
}

func numericArg(prop map[string]any, attempt int) float64 {
	lo, hasLo := bound(prop, "minimum")
	hi, hasHi := bound(prop, "maximum")
	if exclusive, ok := bound(prop, "exclusiveMinimum"); ok {
		lo, hasLo = exclusive+1, true
	}
	var candidates []float64
	if hasHi {
		candidates = append(candidates, hi)
	}
	if hasLo {
		candidates = append(candidates, lo)
	}
	if hasLo && hasHi {
		candidates = append(candidates, (lo+hi)/2)
	}
	candidates = append(candidates, 3600)
	candidates = slices.DeleteFunc(candidates, func(v float64) bool { return (hasLo && v < lo) || (hasHi && v > hi) })
	if len(candidates) == 0 {
		return 1
	}
	return candidates[attempt%len(candidates)]
}

func bound(prop map[string]any, key string) (float64, bool) {
	switch value := prop[key].(type) {
	case float64:
		return value, true
	case json.Number:
		number, err := value.Float64()
		return number, err == nil
	}
	return 0, false
}
