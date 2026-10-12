package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

// Contract names a document's governing contract, as used in validation
// errors: the stem of a schema in docs/contracts or, for the delivery ledger
// and the world-state history (JSON Lines with no schema file), the record
// type the run publishes.
type Contract string

// The contracts whose violations surfaces report with a stable code.
const (
	ContractDomainSpec      Contract = "domain-spec-v0.1"
	ContractOutputAdapter   Contract = "output-adapter-v0.1"
	ContractRunArtifact     Contract = "run-artifact-v0.1"
	ContractConsumerVerdict Contract = "consumer-verdict-v0.1"
	ContractGroundTruth     Contract = "ground-truth-v0.1"
	ContractDeliveryLedger  Contract = "delivery-ledger"
	ContractStateHistory    Contract = "world-state-history"
	ContractScenario        Contract = "scenario-v0.1"
)

// Issue is one way a document fails its contract. A located issue carries an
// RFC 6901 JSON pointer into the document ("" is the whole document); an
// unlocated one names a rule that the validator cannot place.
type Issue struct {
	Pointer string
	Located bool
	Message string
}

// LocatedIssue is an issue at pointer.
func LocatedIssue(pointer, message string) Issue {
	return Issue{Pointer: pointer, Located: true, Message: message}
}

// UnlocatedIssue is an issue whose location in the document is not known.
func UnlocatedIssue(message string) Issue {
	return Issue{Message: message}
}

// MarshalJSON writes the pointer only for a located issue, so "" keeps its
// meaning of "the whole document".
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

// SchemaIssues converts contract-schema violations into located issues.
func SchemaIssues(errs []jsonschema.Error) []Issue {
	issues := make([]Issue, len(errs))
	for i, e := range errs {
		issues[i] = LocatedIssue(e.Path, e.Msg)
	}
	return issues
}

// ValidationError reports a document that fails its contract: the contract,
// the document's source as the caller named it, every issue found, and the
// human summary that Error returns. Cause, when set, is the decoding error
// behind the failure.
type ValidationError struct {
	Contract Contract
	Source   string
	Summary  string
	Issues   []Issue
	Cause    error
}

func (e *ValidationError) Error() string { return e.Summary }

// Unwrap returns the decoding error behind the failure, if any.
func (e *ValidationError) Unwrap() error { return e.Cause }

// NotJSONError reports a document that is not valid JSON. The summary is
// "<prefix>: <cause>"; the single issue is unlocated.
func NotJSONError(contract Contract, source, prefix string, cause error) *ValidationError {
	return &ValidationError{Contract: contract, Source: source, Summary: fmt.Sprintf("%s: %v", prefix, cause),
		Issues: []Issue{UnlocatedIssue(cause.Error())}, Cause: cause}
}

// WithSource names the document behind the first *ValidationError in err's
// chain when it does not name one yet, and returns err. It sets the source in
// place: every validator in this repository returns a fresh error per call,
// so the error is never shared. Any other error is returned unchanged.
func WithSource(err error, source string) error {
	var invalid *ValidationError
	if errors.As(err, &invalid) && invalid.Source == "" {
		invalid.Source = source
	}
	return err
}
