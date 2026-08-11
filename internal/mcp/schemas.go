package mcp

import "encoding/json"

// toolSchema returns the public input contract for one MCP tool. Keep these
// schemas here, next to the MCP surface, so a client can discover the same
// argument names and types that the handlers consume.
func toolSchema(name string) json.RawMessage {
	props := map[string]any{}
	required := []string{}
	var clauses []any
	var allOf []any

	str := func() map[string]any { return map[string]any{"type": "string"} }
	integer := func() map[string]any { return map[string]any{"type": "integer"} }
	boolean := func() map[string]any { return map[string]any{"type": "boolean"} }
	object := func() map[string]any {
		return map[string]any{"type": "object", "additionalProperties": true}
	}
	enum := func(values ...string) map[string]any {
		out := str()
		out["enum"] = values
		return out
	}
	arrayOfStrings := func() map[string]any {
		return map[string]any{"type": "array", "items": str()}
	}
	req := func(names ...string) { required = append(required, names...) }

	switch name {
	case "sim.catalog.list":
		props["group"] = str()
	case "sim.catalog.describe":
		props["domain"] = str()
		req("domain")
	case "sim.catalog.coverage", "sim.adapter.list":
	case "sim.world.create":
		props["domain"] = str()
		props["seed"] = integer()
		props["entities"] = arrayOfStrings()
		props["scenario_profile"] = str()
		props["sink"] = enum("inproc", "file", "http-push")
		props["sink_target"] = str()
		props["adapter"] = str()
		props["time_mode"] = enum("stepped", "scaled", "wall")
		props["start_time"] = integer()
		props["label"] = str()
		req("domain")
		allOf = []any{
			map[string]any{
				"if":   map[string]any{"required": []string{"sink"}, "properties": map[string]any{"sink": map[string]any{"const": "file"}}},
				"then": map[string]any{"required": []string{"sink_target"}},
			},
			map[string]any{
				"if":   map[string]any{"required": []string{"sink"}, "properties": map[string]any{"sink": map[string]any{"const": "http-push"}}},
				"then": map[string]any{"required": []string{"sink_target"}},
			},
		}
	case "sim.world.describe", "sim.world.destroy", "sim.clock.state":
		props["world_id"] = str()
		req("world_id")
	case "sim.clock.advance":
		props["world_id"] = str()
		props["by_ns"] = integer()
		props["to_ns"] = integer()
		props["await_consumer"] = boolean()
		req("world_id")
		clauses = []any{
			map[string]any{"required": []string{"by_ns"}},
			map[string]any{"required": []string{"to_ns"}},
		}
	case "sim.fault.inject":
		props["world_id"] = str()
		props["entity_id"] = str()
		props["fault"] = str()
		props["onset_ns"] = integer()
		props["params"] = object()
		req("world_id", "entity_id", "fault")
	case "sim.fault.clear":
		props["world_id"] = str()
		props["fault_id"] = str()
		req("world_id", "fault_id")
	case "sim.fault.list":
		props["world_id"] = str()
		req("world_id")
	case "sim.perturb.apply":
		props["world_id"] = str()
		props["perturbation"] = str()
		props["params"] = object()
		props["from_ns"] = integer()
		props["until_ns"] = integer()
		req("world_id", "perturbation")
	case "sim.perturb.clear":
		props["world_id"] = str()
		props["perturb_id"] = str()
		req("world_id", "perturb_id")
	case "sim.env.inject":
		props["world_id"] = str()
		props["target"] = str()
		props["fault"] = str()
		props["params"] = object()
		req("world_id", "target", "fault")
	case "sim.run.begin":
		props["world_id"] = str()
		props["label"] = str()
		req("world_id")
	case "sim.run.end":
		props["world_id"] = str()
		req("world_id")
	case "sim.truth.seal":
		props["run_id"] = str()
		props["ground_truth"] = object()
		req("run_id", "ground_truth")
	case "sim.run.verify":
		props["run_artifact_path"] = str()
		req("run_artifact_path")
	case "sim.truth.reveal":
		props["run_id"] = str()
		props["unblind"] = boolean()
		req("run_id")
	case "sim.truth.seal_status", "sim.score":
		props["run_id"] = str()
		req("run_id")
	case "sim.scenario.audit":
		props["domain"] = str()
		props["entity_id"] = str()
		props["fault"] = str()
		props["onset_ns"] = integer()
		props["start_ns"] = integer()
		props["duration_ns"] = integer()
		req("domain", "entity_id", "fault")
	case "sim.entity.retire":
		props["world_id"] = str()
		props["entity_id"] = str()
		props["reason"] = str()
		req("world_id", "entity_id", "reason")
	case "sim.nameplate.read", "sim.effector.list":
		props["token"] = str()
		req("token")
	case "sim.effector.invoke":
		props["token"] = str()
		props["effector"] = str()
		props["entity_id"] = str()
		props["command_id"] = str()
		props["args"] = object()
		props["at_ns"] = integer()
		req("token", "effector", "entity_id", "command_id")
	case "sim.consumer.report":
		props["token"] = str()
		props["run_id"] = str()
		props["quiesced_through_ns"] = integer()
		props["verdict"] = object()
		req("token", "run_id")
	default:
		panic("mcp: schema is missing for tool " + name)
	}

	doc := map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             required,
		"additionalProperties": false,
	}
	if len(clauses) > 0 {
		doc["oneOf"] = clauses
	}
	if len(allOf) > 0 {
		doc["allOf"] = allOf
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		panic("mcp: marshal tool schema: " + err.Error())
	}
	return raw
}
