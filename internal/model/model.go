// Package model holds the typed forms of every simulator contract: the
// domain spec, the native sim event, the output adapter, the consumer
// verdict, the ground-truth record, the run artifact, and the delivery
// ledger. These types mirror docs/contracts/*.schema.json field for field.
// Domain and adapter documents are validated by their own packages; this
// package validates only the consumer verdict and the run artifact, the two
// contracts that cross the run boundary.
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

// SimVersion is part of the determinism tuple. A different simulator version
// may legitimately produce a different trace; verify reports that as a
// version mismatch rather than a failure.
const SimVersion = "0.1.0"

// DefaultStartTimeNS is the epoch used when a world is created without an
// explicit start time: 2026-01-01T00:00:00Z.
const DefaultStartTimeNS int64 = 1767225600000000000

// FormatTime renders a nanosecond epoch as RFC 3339 with nanoseconds, UTC.
func FormatTime(ns int64) string {
	return time.Unix(0, ns).UTC().Format(time.RFC3339Nano)
}

// ParseTime parses an RFC 3339 timestamp into a nanosecond epoch.
func ParseTime(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0, fmt.Errorf("model: invalid date-time %q: %w", s, err)
	}
	return t.UnixNano(), nil
}

// Decode decodes JSON from r into dst, preserving number literals as
// json.Number inside map[string]any values so that timestamps and seeds
// beyond float64 precision survive.
func Decode(r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("model: %w", err)
	}
	return rejectTrailingJSON(dec)
}

// DecodeBytes is Decode over a byte slice.
func DecodeBytes(b []byte, dst any) error {
	return Decode(strings.NewReader(string(b)), dst)
}

// ValidateVerdict checks a serialized consumer verdict against the
// committed consumer-verdict schema.
func ValidateVerdict(raw []byte) error {
	return validateAgainst(raw, "verdict", "consumer-verdict-v0.1", schemas.ConsumerVerdict())
}

// ValidateRunArtifact checks a serialized run artifact against the
// committed run-artifact schema.
func ValidateRunArtifact(raw []byte) error {
	return validateAgainst(raw, "artifact", "run-artifact-v0.1", schemas.RunArtifact())
}

func validateAgainst(raw []byte, noun, contract string, schema []byte) error {
	var doc any
	if err := DecodeBytes(raw, &doc); err != nil {
		return fmt.Errorf("model: %s not valid JSON: %w", noun, err)
	}
	sch, err := jsonschema.CompileJSON(schema)
	if err != nil {
		return fmt.Errorf("model: compile %s schema: %w", noun, err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return fmt.Errorf("model: %s fails %s: %s", noun, contract, errs[0].Error())
	}
	return nil
}

func rejectTrailingJSON(dec *json.Decoder) error {
	// Reject trailing garbage after the document.
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("model: trailing JSON document")
		}
		return fmt.Errorf("model: trailing data after JSON document: %w", err)
	}
	return nil
}
