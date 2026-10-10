package protocol

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func tokenInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"token": stringSchema(),
		},
		Required: []string{"token"},
	}
}

func effectorInvokeInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"token":      stringSchema(),
			"effector":   stringSchema(),
			"entity_id":  stringSchema(),
			"command_id": stringSchema(),
			"args":       objectSchema(),
			"at_ns":      integerSchema(),
		},
		Required: []string{"token", "effector", "entity_id", "command_id"},
	}
}

func consumerReportInput() toolInput {
	return toolInput{
		Properties: map[string]any{
			"token":               stringSchema(),
			"run_id":              stringSchema(),
			"quiesced_through_ns": integerSchema(),
			"verdict":             embeddedContract(schemas.ConsumerVerdict()),
		},
		Required: []string{"token", "run_id"},
	}
}
