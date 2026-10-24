package protocol

func jobStartInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"domain":     stringSchema(),
			"seed":       seedSchema(),
			"explore":    booleanSchema(),
			"duration_s": map[string]any{"type": "integer", "minimum": 1},
			"rate":       map[string]any{"type": "number", "minimum": 0},
			"faults":     map[string]any{"type": "array", "items": jobFaultSchema()},
			"calls":      map[string]any{"type": "array", "items": jobCallSchema()},
		},
		Required: []string{"domain", "seed", "duration_s"},
	}
}

func jobFaultSchema() map[string]any {
	return closedObject(map[string]any{
		"entity_id": stringSchema(),
		"fault":     stringSchema(),
		"at_s":      map[string]any{"type": "integer", "minimum": 0},
	}, "entity_id", "fault", "at_s")
}

func jobCallSchema() map[string]any {
	return closedObject(map[string]any{
		"entity_id": stringSchema(),
		"effector":  stringSchema(),
		"args":      objectSchema(),
		"at_s":      map[string]any{"type": "integer", "minimum": 0},
	}, "entity_id", "effector", "at_s")
}

func closedObject(properties map[string]any, required ...string) map[string]any {
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}

func jobRateInput() toolInput {
	return toolInput{
		Properties: map[string]any{"world_id": stringSchema(), "rate": map[string]any{"type": "number", "minimum": 0}},
		Required:   []string{"world_id", "rate"},
	}
}

func jobInjectFaultInput() toolInput {
	return toolInput{
		Properties: map[string]any{"world_id": stringSchema(), "entity_id": stringSchema(), "fault": stringSchema()},
		Required:   []string{"world_id", "entity_id", "fault"},
	}
}

func jobInvokeEffectorInput() toolInput {
	return toolInput{
		Properties: map[string]any{"world_id": stringSchema(), "entity_id": stringSchema(), "effector": stringSchema(), "args": objectSchema()},
		Required:   []string{"world_id", "entity_id", "effector"},
	}
}
