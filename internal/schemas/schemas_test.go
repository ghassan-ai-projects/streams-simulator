package schemas

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEmbeddedMatchCommitted proves the embedded schemas never drift from the
// committed originals in docs/contracts.
func TestEmbeddedMatchCommitted(t *testing.T) {
	cases := []struct {
		name string
		got  []byte
	}{
		{"domain-spec-v0.1.schema.json", DomainSpec()},
		{"sim-event-v0.1.schema.json", SimEvent()},
		{"output-adapter-v0.1.schema.json", OutputAdapter()},
		{"consumer-verdict-v0.1.schema.json", ConsumerVerdict()},
		{"ground-truth-v0.1.schema.json", GroundTruth()},
		{"run-artifact-v0.1.schema.json", RunArtifact()},
	}
	for _, tc := range cases {
		want, err := os.ReadFile(filepath.Join("..", "..", "docs", "contracts", tc.name))
		if err != nil {
			t.Fatalf("read committed schema %s: %v", tc.name, err)
		}
		if string(tc.got) != string(want) {
			t.Errorf("embedded schema %s drifted from docs/contracts: re-run the copy step", tc.name)
		}
	}
}
