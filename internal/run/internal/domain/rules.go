// Package domain holds the run module's pure rules: identity and digest
// derivation, artifact document validation and decoding, replay command
// argument decoding and verdict validation. It performs no I/O and reads no
// clock.
package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// errTrailingDocument is the cause of an artifact followed by another JSON value.
var errTrailingDocument = errors.New("a second JSON value follows the artifact")

// ValidateArtifactDocument checks that raw is a JSON document that satisfies
// the run-artifact schema.
func ValidateArtifactDocument(raw []byte, path string) error {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return model.NotJSONError(model.ContractRunArtifact, path, fmt.Sprintf("run: %s not valid JSON", path), err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	return nil
}

// DecodeArtifact decodes a validated artifact document, rejecting trailing
// JSON.
func DecodeArtifact(raw []byte) (*model.RunArtifact, error) {
	var artifact model.RunArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&artifact); err != nil {
		return nil, model.NotJSONError(model.ContractRunArtifact, "", "run: decode artifact", err)
	}
	if err := rejectArtifactTrailingJSON(decoder); err != nil {
		return nil, err
	}
	return &artifact, nil
}

func rejectArtifactTrailingJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return model.NotJSONError(model.ContractRunArtifact, "", "run: artifact has trailing JSON", errTrailingDocument)
		}
		return model.NotJSONError(model.ContractRunArtifact, "", "run: artifact trailing data", err)
	}
	return nil
}

// ErrorString renders an error, or "" for nil.
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// CountLedger counts the ledger rows whose delivery reason is one of reasons.
func CountLedger(ledger []model.LedgerRecord, reasons ...string) int {
	n := 0
	for _, l := range ledger {
		for _, r := range reasons {
			if l.DeliveryReason == r {
				n++
				break
			}
		}
	}
	return n
}

func RunHash(domainID string) uint64 {
	b, _ := canonical.MarshalString(map[string]any{"d": domainID})
	return fnv([]byte(b))
}

func fnv(b []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range b {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

// AdapterDigest is the RFC 8785 canonical digest of an adapter's JSON form,
// so it does not depend on struct field order or on how the source file was
// formatted. (An explicit zero kept out by `omitempty` is not part of the
// form: the digest identifies the adapter as the simulator understands it.)
func AdapterDigest(a *model.Adapter) string {
	raw, err := json.Marshal(a)
	if err != nil {
		return "invalid-adapter-digest:" + err.Error()
	}
	var form any
	if err := model.DecodeBytes(raw, &form); err != nil {
		return "invalid-adapter-digest:" + err.Error()
	}
	digest, err := canonical.Digest(form)
	if err != nil {
		return "invalid-adapter-digest:" + err.Error()
	}
	return digest
}

// ValidateSubmittedVerdict refuses a verdict the run cannot accept.
func ValidateSubmittedVerdict(v *model.Verdict) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	if err := model.ValidateVerdict(raw); err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	return nil
}

// CloneVerdict deep-copies a verdict so callers cannot mutate the run's copy.
func CloneVerdict(v *model.Verdict) *model.Verdict {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	out := &model.Verdict{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil
	}
	return out
}

// CommandString reads a string argument of a recorded command, "" when absent.
func CommandString(args map[string]any, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

// CommandTime reads a nanosecond time argument of a recorded command, 0 when
// it is absent or not an exact time (see ExactTime).
func CommandTime(args map[string]any, key string) int64 {
	value, _ := ExactTime(args, key)
	return value
}

// ExactTime reads a nanosecond time argument, reporting false unless it is an
// integer that fits in 64 bits. A decoded document's time must be written as
// a plain integer: a fraction or an exponent is not an exact time.
func ExactTime(args map[string]any, key string) (int64, bool) {
	switch value := args[key].(type) {
	case int64:
		return value, true
	case float64:
		return int64(value), value == math.Trunc(value) && math.Abs(value) < math.MaxInt64
	case json.Number:
		n, err := value.Int64()
		return n, err == nil
	}
	return 0, false
}

// AsMap views a decoded JSON value as an object, nil when it is not one.
func AsMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
