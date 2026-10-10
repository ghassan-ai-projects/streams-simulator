package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// A profile weighting several undeclared faults is reported by the first
// name in sorted order, every time: validation does not depend on map order.
func TestProfileValidationNamesTheSortedFirstUndeclaredFault(t *testing.T) {
	t.Parallel()
	c, err := Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(c.Spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	profile := doc["profiles"].([]any)[0].(map[string]any)
	profile["fault_weights"] = map[string]any{"zz_unknown": 1.0, "aa_unknown": 1.0, "mm_unknown": 1.0}
	broken, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		if _, err := Parse(broken, "broken"); err == nil || !strings.Contains(err.Error(), `undeclared fault "aa_unknown"`) {
			t.Fatalf("err = %v, want the sorted-first undeclared fault", err)
		}
	}
}
