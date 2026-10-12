// Package model holds the typed forms of every simulator contract: the
// domain spec, the native sim event, the output adapter, the consumer
// verdict, the ground-truth record, the run artifact, and the delivery
// ledger. These types mirror docs/contracts/*.schema.json field for field.
// Domain and adapter documents are validated by their own packages; this
// package validates the consumer verdict, the ground-truth record and the run
// artifact, the contracts that cross the run boundary. It also holds the
// shared validation-error record (ValidationError, Issue, Contract) every
// module reports a contract violation with.
package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

const SimVersion = "0.3.0"

const DefaultStartTimeNS int64 = 1767225600000000000

func FormatTime(ns int64) string {
	return time.Unix(0, ns).UTC().Format(time.RFC3339Nano)
}

func ParseTime(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0, fmt.Errorf("model: invalid date-time %q: %w", s, err)
	}
	return t.UnixNano(), nil
}

func Decode(r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("model: %w", err)
	}
	return rejectTrailingJSON(dec)
}

func DecodeBytes(b []byte, dst any) error {
	return Decode(strings.NewReader(string(b)), dst)
}

func ValidateVerdict(raw []byte) error {
	return validateAgainst(raw, "verdict", ContractConsumerVerdict, schemas.ConsumerVerdict())
}

func ValidateGroundTruth(raw []byte) error {
	return validateAgainst(raw, "ground truth", ContractGroundTruth, schemas.GroundTruth())
}

func ValidateRunArtifact(raw []byte) error {
	return validateAgainst(raw, "artifact", ContractRunArtifact, schemas.RunArtifact())
}

func validateAgainst(raw []byte, noun string, contract Contract, schema []byte) error {
	var doc any
	if err := DecodeBytes(raw, &doc); err != nil {
		return NotJSONError(contract, "", fmt.Sprintf("model: %s not valid JSON", noun), err)
	}
	sch, err := jsonschema.CompileJSON(schema)
	if err != nil {
		return fmt.Errorf("model: compile %s schema: %w", noun, err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		summary := fmt.Sprintf("model: %s fails %s: %s", noun, contract, errs[0].Error())
		return &ValidationError{Contract: contract, Summary: summary, Issues: SchemaIssues(errs)}
	}
	return nil
}

func rejectTrailingJSON(dec *json.Decoder) error {

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("model: trailing JSON document")
		}
		return fmt.Errorf("model: trailing data after JSON document: %w", err)
	}
	return nil
}
