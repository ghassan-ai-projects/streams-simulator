package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

const ScenarioSchemaVersion = "0.1"

type Scenario struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Description   string            `json:"description,omitempty"`
	Domain        string            `json:"domain"`
	Adapter       string            `json:"adapter"`
	Seed          uint64            `json:"seed"`
	World         ScenarioWorld     `json:"world"`
	Commands      []ScenarioCommand `json:"commands"`
}

type ScenarioWorld struct {
	StartTimeNS      int64    `json:"start_time_ns"`
	Profile          string   `json:"profile,omitempty"`
	EntityIDs        []string `json:"entity_ids,omitempty"`
	Noiseless        bool     `json:"noiseless,omitempty"`
	ForceFailureMode string   `json:"force_failure_mode,omitempty"`
	ClockMultiplier  float64  `json:"clock_multiplier,omitempty"`
}

type ScenarioCommand struct {
	Op   string         `json:"op"`
	Args map[string]any `json:"args"`
}

func ValidateScenario(raw []byte) error {
	return validateAgainst(raw, "scenario", ContractScenario, schemas.Scenario())
}

func DecodeScenario(raw []byte) (*Scenario, error) {
	if err := ValidateScenario(raw); err != nil {
		return nil, err
	}
	var scenario Scenario
	if err := DecodeBytes(raw, &scenario); err != nil {
		return nil, scenarioDecodeError(err)
	}
	return &scenario, nil
}

func scenarioDecodeError(err error) *ValidationError {
	issue := UnlocatedIssue(err.Error())
	var mismatch *json.UnmarshalTypeError
	if errors.As(err, &mismatch) && mismatch.Field != "" {
		issue = LocatedIssue("/"+strings.ReplaceAll(mismatch.Field, ".", "/"), fmt.Sprintf("%s does not fit a %s", mismatch.Value, mismatch.Type))
	}
	summary := "model: scenario cannot be decoded: " + issue.Message
	return &ValidationError{Contract: ContractScenario, Summary: summary, Issues: []Issue{issue}, Cause: err}
}
