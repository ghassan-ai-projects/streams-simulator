package protocol

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func catalogListInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"group": stringSchema(),
		},
	}
}

func catalogDescribeInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"domain": stringSchema(),
		},
		Required: []string{"domain"},
	}
}

func emptyToolInput() toolInput {
	return toolInput{
		Properties: map[string]any{},
	}
}

func worldCreateInput() toolInput {
	return toolInput{
		Properties: worldCreationProperties(),
		Required:   []string{"domain"},
		AllOf:      sinkTargetConditions(),
	}
}

func worldIDInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id": stringSchema(),
		},
		Required: []string{"world_id"},
	}
}

func clockAdvanceInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id":       stringSchema(),
			"by_ns":          integerSchema(),
			"to_ns":          integerSchema(),
			"await_consumer": booleanSchema(),
		},
		Required: []string{"world_id"},
		OneOf:    clockAdvanceChoice(),
	}
}

func faultInjectInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id":  stringSchema(),
			"entity_id": stringSchema(),
			"fault":     stringSchema(),
			"onset_ns":  integerSchema(),
			"params":    objectSchema(),
		},
		Required: []string{"world_id", "entity_id", "fault"},
	}
}

func faultClearInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id": stringSchema(),
			"fault_id": stringSchema(),
		},
		Required: []string{"world_id", "fault_id"},
	}
}

func perturbApplyInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id":     stringSchema(),
			"perturbation": stringSchema(),
			"params":       objectSchema(),
			"from_ns":      integerSchema(),
			"until_ns":     integerSchema(),
		},
		Required: []string{"world_id", "perturbation"},
	}
}

func perturbClearInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id":   stringSchema(),
			"perturb_id": stringSchema(),
		},
		Required: []string{"world_id", "perturb_id"},
	}
}

func envInjectInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id": stringSchema(),
			"target":   stringSchema(),
			"fault":    stringSchema(),
			"params":   objectSchema(),
		},
		Required: []string{"world_id", "target", "fault"},
	}
}

func runBeginInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id": stringSchema(),
			"label":    stringSchema(),
		},
		Required: []string{"world_id"},
	}
}

func truthSealInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"run_id":       stringSchema(),
			"ground_truth": embeddedContract(schemas.GroundTruth()),
		},
		Required: []string{"run_id", "ground_truth"},
	}
}

func runVerifyInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"run_artifact_path": stringSchema(),
		},
		Required: []string{"run_artifact_path"},
	}
}

func truthRevealInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"run_id":  stringSchema(),
			"unblind": booleanSchema(),
		},
		Required: []string{"run_id"},
	}
}

func runIDInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"run_id": stringSchema(),
		},
		Required: []string{"run_id"},
	}
}

func scenarioAuditInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"domain":      stringSchema(),
			"entity_id":   stringSchema(),
			"fault":       stringSchema(),
			"onset_ns":    integerSchema(),
			"start_ns":    integerSchema(),
			"duration_ns": integerSchema(),
		},
		Required: []string{"domain", "entity_id", "fault"},
	}
}

func entityRetireInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"world_id":  stringSchema(),
			"entity_id": stringSchema(),
			"reason":    stringSchema(),
		},
		Required: []string{"world_id", "entity_id", "reason"},
	}
}
