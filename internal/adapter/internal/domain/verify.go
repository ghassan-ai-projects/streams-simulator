package domain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type VerifyResult struct {
	Adapter string `json:"adapter"`

	SchemaChecked   bool   `json:"schema_checked"`
	SchemaOK        bool   `json:"schema_ok"`
	GoldenChecked   bool   `json:"golden_checked"`
	GoldenMatch     bool   `json:"golden_match"`
	RecordCount     int    `json:"record_count"`
	FirstDivergence string `json:"first_divergence,omitempty"`
	Detail          string `json:"detail,omitempty"`
}

type VerifyFiles interface {
	OutputSchema(path string) ([]byte, error)
	Golden(path string) ([]byte, error)
}

func Verify(a *model.Adapter, fixture []model.SimEvent, base string, src VerifyFiles) (*VerifyResult, error) {
	if len(fixture) == 0 {
		return nil, fmt.Errorf("adapter: the conformance fixture holds no events")
	}
	if err := validateStrictObservedOrder(fixture); err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	out, err := renderVerification(a, fixture)
	if err != nil {
		return nil, err
	}
	return verifyRendered(a, out, base, src)
}

func renderVerification(a *model.Adapter, fixture []model.SimEvent) ([]byte, error) {
	engine, err := NewEngine(a, verificationMetadata(fixture))
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	out, err := engine.RenderRun(fixture, mustParseTime(fixture[len(fixture)-1].ObservedTime))
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return out, nil
}

func verificationMetadata(fixture []model.SimEvent) map[string]any {
	return map[string]any{
		"run_id": "verify", "sim_version": "0.1.0",
		"domain_id": "fixture", "domain_version": "0.0.0",
		"world_start_time": fixture[0].EventTime, "world_end_time": fixture[len(fixture)-1].EventTime,
		"seed": float64(0),
	}
}

func verifyRendered(a *model.Adapter, out []byte, base string, src VerifyFiles) (*VerifyResult, error) {
	res := &VerifyResult{Adapter: a.ID}
	if complete, err := verifyOutputSchema(a, out, base, src, res); err != nil {
		return nil, err
	} else if !complete {
		return res, nil
	}
	if err := verifyGolden(a, out, base, src, res); err != nil {
		return nil, err
	}
	return res, nil
}

func verifyOutputSchema(a *model.Adapter, out []byte, base string, src VerifyFiles, res *VerifyResult) (bool, error) {
	if a.Conformance == nil || a.Conformance.OutputSchema == "" {
		return true, nil
	}
	res.SchemaChecked = true
	path := resolveConformancePath(base, a.Conformance.OutputSchema)
	schema, err := loadOutputSchema(path, src)
	if err != nil {
		return false, err
	}
	return validateRenderedRecords(a, out, schema, filepath.Base(path), res)
}

func validateRenderedRecords(a *model.Adapter, out []byte, schema *jsonschema.Schema, name string, res *VerifyResult) (bool, error) {
	records, err := splitRecords(out, a.Encoding)
	if err != nil {
		return false, fmt.Errorf("adapter: %w", err)
	}
	return validateOutputRecords(records, schema, name, res), nil
}

func verifyGolden(a *model.Adapter, out []byte, base string, src VerifyFiles, res *VerifyResult) error {
	if a.Conformance == nil || a.Conformance.Golden == "" {
		return nil
	}
	res.GoldenChecked = true
	return compareGolden(resolveConformancePath(base, a.Conformance.Golden), out, src, res)
}

func validateStrictObservedOrder(events []model.SimEvent) error {
	var previous int64
	for i, ev := range events {
		observed, err := model.ParseTime(ev.ObservedTime)
		if err != nil {
			return fmt.Errorf("strict observed-time order: event %d: %w", i, err)
		}
		if i > 0 && observed <= previous {
			return fmt.Errorf("strict observed-time order: event %d observed_time %s is not after %s", i, ev.ObservedTime, model.FormatTime(previous))
		}
		previous = observed
	}
	return nil
}

func resolveConformancePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if base == "" {
		return p
	}
	return filepath.Join(base, p)
}

func splitRecords(out []byte, encoding string) ([]string, error) {
	if encoding == "json-array" {
		return arrayRecords(out)
	}
	return lineRecords(out), nil
}

func arrayRecords(out []byte) ([]string, error) {
	var elements []json.RawMessage
	if err := json.Unmarshal(out, &elements); err != nil {
		return nil, fmt.Errorf("json-array output is not a JSON array: %w", err)
	}
	records := make([]string, len(elements))
	for i, element := range elements {
		records[i] = string(element)
	}
	return records, nil
}

func lineRecords(out []byte) []string {
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	var recs []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		recs = append(recs, l)
	}
	return recs
}

func mustParseTime(ts string) int64 {
	n, _ := model.ParseTime(ts)
	return n
}
