package protocol

import "encoding/json"

func stringSchema() map[string]any  { return map[string]any{"type": "string"} }
func integerSchema() map[string]any { return map[string]any{"type": "integer"} }

// maxMCPSeed is the largest seed MCP can carry exactly: JSON numbers reach the
// handler as float64, and every integer up to 2^53-1 is exact in float64.
// A larger seed would be rounded without notice, so it is refused instead.
const maxMCPSeed = 1<<53 - 1

func seedSchema() map[string]any {
	return map[string]any{"type": "integer", "minimum": 0, "maximum": maxMCPSeed}
}

func booleanSchema() map[string]any { return map[string]any{"type": "boolean"} }
func objectSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true}
}
func enumSchema(values ...string) map[string]any {
	out := stringSchema()
	out["enum"] = values
	return out
}
func stringArraySchema() map[string]any {
	return map[string]any{"type": "array", "items": stringSchema()}
}

// Embedded contracts omit document annotations when used as nested arguments.
// The SDK still enforces the exact committed validation keywords over the wire.
func embeddedContract(contract []byte) map[string]any {
	doc := map[string]any{}
	if err := json.Unmarshal(contract, &doc); err != nil {
		panic("mcp: unmarshal embedded contract schema: " + err.Error())
	}
	delete(doc, "$schema")
	delete(doc, "$id")
	delete(doc, "title")
	delete(doc, "description")
	return doc
}

func sinkTargetConditions() []any {
	return []any{
		map[string]any{
			"if":   map[string]any{"required": []string{"sink"}, "properties": map[string]any{"sink": map[string]any{"const": "file"}}},
			"then": map[string]any{"required": []string{"sink_target"}},
		},
		map[string]any{
			"if":   map[string]any{"required": []string{"sink"}, "properties": map[string]any{"sink": map[string]any{"const": "http-push"}}},
			"then": map[string]any{"required": []string{"sink_target"}},
		},
	}
}

func clockAdvanceChoice() []any {
	return []any{
		map[string]any{"required": []string{"by_ns"}},
		map[string]any{"required": []string{"to_ns"}},
	}
}

func worldCreationProperties() map[string]any {
	return map[string]any{
		"domain":           stringSchema(),
		"seed":             seedSchema(),
		"entities":         stringArraySchema(),
		"scenario_profile": stringSchema(),
		"sink":             enumSchema("inproc", "file", "http-push"),
		"sink_target":      stringSchema(),
		"adapter":          stringSchema(),
		"time_mode":        enumSchema("stepped", "scaled", "wall"),
		"start_time":       integerSchema(),
		"label":            stringSchema(),
	}
}
