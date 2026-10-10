// Package conformance proves an output adapter correct with the consumer
// absent: it renders the shipped fixture through the adapter engine, validates
// the output against the adapter's declared schema and byte-compares it to the
// committed golden file. It is the `adapter verify` use case.
package conformance

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Result is the outcome of `streamsim adapter verify`.
type Result struct {
	Adapter         string `json:"adapter"`
	SchemaOK        bool   `json:"schema_ok"`
	GoldenMatch     bool   `json:"golden_match"`
	RecordCount     int    `json:"record_count"`
	FirstDivergence string `json:"first_divergence,omitempty"`
	Detail          string `json:"detail,omitempty"`
}

// Verify proves an adapter correct with the consumer absent: it renders the
// shipped 12-event fixture, validates every rendered record against the
// adapter's declared output schema, and byte-compares the full output to the
// committed golden file. adapterPath is the adapter file; fixturePath is the
// native-event fixture (defaults to the embedded one); base is the directory
// conformance paths resolve against.
func Verify(adapterPath, fixturePath, base string) (*Result, error) {
	a, fixture, err := verificationInputs(adapterPath, fixturePath)
	if err != nil {
		return nil, err
	}
	out, err := renderVerification(a, fixture)
	if err != nil {
		return nil, err
	}
	return verifyRendered(a, out, base)
}

func verifyOutputSchema(a *model.Adapter, out []byte, base string, res *Result) (bool, error) {
	if a.Conformance == nil || a.Conformance.OutputSchema == "" {
		return true, nil
	}
	path := resolvePath(base, a.Conformance.OutputSchema)
	schema, err := loadOutputSchema(path)
	if err != nil {
		return false, err
	}
	records, err := splitRecords(out, a.Encoding)
	if err != nil {
		return false, fmt.Errorf("adapter: %w", err)
	}
	return validateOutputRecords(records, schema, filepath.Base(path), res), nil
}

func verifyGolden(a *model.Adapter, out []byte, base string, res *Result) error {
	if a.Conformance == nil || a.Conformance.Golden == "" {
		return nil
	}
	return compareGolden(resolvePath(base, a.Conformance.Golden), out, res)
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

func firstDivergence(got, want []byte) string {
	line, col := 1, 1
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			return fmt.Sprintf("first divergent byte at line %d, column %d", line, col)
		}
		advanceTextPosition(&line, &col, got[i])
	}
	return "lengths differ"
}

func resolvePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if base == "" {
		return p
	}
	return filepath.Join(base, p)
}

func splitRecords(out []byte, encoding string) ([]string, error) {
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	var recs []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		recs = append(recs, l)
	}
	return recs, nil
}

func mustParse(ts string) int64 {
	n, _ := model.ParseTime(ts)
	return n
}

func advanceTextPosition(line, col *int, value byte) {
	if value == '\n' {
		*line++
		*col = 1
	} else {
		*col++
	}
}
