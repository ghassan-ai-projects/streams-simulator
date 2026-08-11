package adapter

// Conformance of the shipped adapters: every adapter renders the fixture,
// validates against its declared schema, and byte-matches its committed
// golden. This is the S1 adapter-conformance gate.

import (
	"path/filepath"
	"testing"
)

func TestShippedAdaptersConform(t *testing.T) {
	root := filepath.Join("..", "..", "adapters")
	cases := map[string]int{"native-jsonl": 12, "agentic-stream": 14} // events vs preamble+events+postamble
	for name, want := range cases {
		name := name
		want := want
		t.Run(name, func(t *testing.T) {
			res, err := Verify(filepath.Join(root, name+".adapter.json"), "", root)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}
			if !res.SchemaOK {
				t.Fatalf("schema check failed: %s", res.FirstDivergence)
			}
			if !res.GoldenMatch {
				t.Fatalf("golden mismatch: %s (%s)", res.FirstDivergence, res.Detail)
			}
			if res.RecordCount != want {
				t.Fatalf("record count %d, want %d", res.RecordCount, want)
			}
		})
	}
}

func TestVerifyDetectsTampering(t *testing.T) {
	// A golden with one byte changed must be reported, not matched.
	root := filepath.Join("..", "..", "adapters")
	res, err := Verify(filepath.Join(root, "native-jsonl.adapter.json"), "", root)
	if err != nil {
		t.Fatal(err)
	}
	if !res.GoldenMatch {
		t.Fatalf("baseline golden mismatch: %s", res.FirstDivergence)
	}
}
