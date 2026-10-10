// Package domain holds the run module's pure rules: identity and digest
// derivation, artifact document validation and decoding, replay command
// argument decoding and verdict validation. It performs no I/O and reads no
// clock.
package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// ValidateArtifactDocument checks that raw is a JSON document that satisfies
// the run-artifact schema.
func ValidateArtifactDocument(raw []byte, path string) error {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return fmt.Errorf("run: %s not valid JSON: %w", path, err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		return fmt.Errorf("streamsim: %w", err)
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
		return nil, fmt.Errorf("run: decode artifact: %w", err)
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
			return fmt.Errorf("run: artifact has trailing JSON")
		}
		return fmt.Errorf("run: artifact trailing data: %w", err)
	}
	return nil
}

func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

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

// canonicalHash derives a short stable id from the domain and seed.
func CanonicalHash(domainID string, seed uint64) uint64 {
	b, _ := canonical.MarshalString(map[string]any{"d": domainID, "s": seed})
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

func AdapterDigest(a *model.Adapter) string {
	raw, _ := json.Marshal(a)
	return canonical.DigestBytes(raw)
}

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

func CommandString(args map[string]any, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

func CommandTime(args map[string]any, key string) int64 {
	switch value := args[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case json.Number:
		if n, err := value.Int64(); err == nil {
			return n
		}
	}
	return 0
}

func AsMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
