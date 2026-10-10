package domain

// Conformance verification (`adapter verify`): render the shipped fixture
// through the adapter, validate every record against the adapter's declared
// output schema and byte-compare the whole output to the committed golden.
// The files verification compares against arrive through VerifyFiles.

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// VerifyResult is the outcome of `streamsim adapter verify`.
type VerifyResult struct {
	Adapter         string `json:"adapter"`
	SchemaOK        bool   `json:"schema_ok"`
	GoldenMatch     bool   `json:"golden_match"`
	RecordCount     int    `json:"record_count"`
	FirstDivergence string `json:"first_divergence,omitempty"`
	Detail          string `json:"detail,omitempty"`
}

// VerifyFiles reads the files a verification compares against, by the paths
// the adapter declares (already resolved against the base directory).
type VerifyFiles interface {
	OutputSchema(path string) ([]byte, error)
	Golden(path string) ([]byte, error)
}

// Verify proves an adapter correct with the consumer absent: it renders the
// fixture, validates every rendered record against the adapter's declared
// output schema, and byte-compares the full output to the committed golden
// file. base is the directory the adapter's conformance paths resolve
// against.
func Verify(a *model.Adapter, fixture []model.SimEvent, base string, src VerifyFiles) (*VerifyResult, error) {
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
	path := resolveConformancePath(base, a.Conformance.OutputSchema)
	schema, err := loadOutputSchema(path, src)
	if err != nil {
		return false, err
	}
	records := splitRecords(out)
	return validateOutputRecords(records, schema, filepath.Base(path), res), nil
}

func verifyGolden(a *model.Adapter, out []byte, base string, src VerifyFiles, res *VerifyResult) error {
	if a.Conformance == nil || a.Conformance.Golden == "" {
		return nil
	}
	return compareGolden(resolveConformancePath(base, a.Conformance.Golden), out, src, res)
}

// validateStrictObservedOrder protects the adapter conformance path from a
// fixture that would violate the simulator's native single-stream ordering
// guarantee. Deliberate delivery perturbations are applied after this source
// contract and remain available to exercise consumers' rejection paths.
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

func splitRecords(out []byte) []string {
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
