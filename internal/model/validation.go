package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

type Contract string

const (
	ContractDomainSpec      Contract = "domain-spec-v0.2"
	ContractOutputAdapter   Contract = "output-adapter-v0.1"
	ContractRunArtifact     Contract = "run-artifact-v0.1"
	ContractConsumerVerdict Contract = "consumer-verdict-v0.1"
	ContractGroundTruth     Contract = "ground-truth-v0.1"
	ContractDeliveryLedger  Contract = "delivery-ledger"
	ContractStateHistory    Contract = "world-state-history"
	ContractScenario        Contract = "scenario-v0.1"
	ContractComponent       Contract = "component-v0.1"
	ContractInstrument      Contract = "instrument-v0.1"
	ContractComposition     Contract = "composition-v0.1"
	ContractVocabulary      Contract = "component-vocabulary-v0.1"
	ContractKindFile        Contract = "kind-file-v0.1"
)

type Issue struct {
	Pointer string
	Located bool
	Message string
}

func LocatedIssue(pointer, message string) Issue {
	return Issue{Pointer: pointer, Located: true, Message: message}
}

func UnlocatedIssue(message string) Issue {
	return Issue{Message: message}
}

func (i Issue) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(i.wire())
	if err != nil {
		return nil, fmt.Errorf("model: issue: %w", err)
	}
	return raw, nil
}

func (i Issue) wire() any {
	if !i.Located {
		return struct {
			Message string `json:"message"`
		}{i.Message}
	}
	return struct {
		Pointer string `json:"pointer"`
		Message string `json:"message"`
	}{i.Pointer, i.Message}
}

func SchemaIssues(errs []jsonschema.Error) []Issue {
	issues := make([]Issue, len(errs))
	for i, e := range errs {
		issues[i] = LocatedIssue(e.Path, e.Msg)
	}
	return issues
}

type ValidationError struct {
	Contract Contract
	Source   string
	Summary  string
	Issues   []Issue
	Cause    error
}

func (e *ValidationError) Error() string { return e.Summary }

func (e *ValidationError) Unwrap() error { return e.Cause }

func NotJSONError(contract Contract, source, prefix string, cause error) *ValidationError {
	return &ValidationError{Contract: contract, Source: source, Summary: fmt.Sprintf("%s: %v", prefix, cause),
		Issues: []Issue{UnlocatedIssue(cause.Error())}, Cause: cause}
}

func WithSource(err error, source string) error {
	var invalid *ValidationError
	if errors.As(err, &invalid) && invalid.Source == "" {
		invalid.Source = source
	}
	return err
}
