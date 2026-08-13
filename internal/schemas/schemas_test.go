package schemas

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
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

// TestEmbeddedDomainSchemaValidatesAllShippedDomains prevents a schema update
// from being accepted without validating the catalog data it is meant to
// describe. In particular, profile setup calls must remain part of the
// domain-spec contract because the suite consumes them at runtime.
func TestEmbeddedDomainSchemaValidatesAllShippedDomains(t *testing.T) {
	var schemaDoc any
	dec := json.NewDecoder(bytes.NewReader(DomainSpec()))
	dec.UseNumber()
	if err := dec.Decode(&schemaDoc); err != nil {
		t.Fatalf("decode embedded domain schema: %v", err)
	}
	schema, err := jsonschema.Compile(schemaDoc)
	if err != nil {
		t.Fatalf("compile embedded domain schema: %v", err)
	}

	domainsDir := filepath.Join("..", "..", "domains")
	entries, err := os.ReadDir(domainsDir)
	if err != nil {
		t.Fatalf("read shipped domains: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		count++
		path := filepath.Join(domainsDir, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read domain: %v", err)
			}
			var doc any
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.UseNumber()
			if err := dec.Decode(&doc); err != nil {
				t.Fatalf("decode domain: %v", err)
			}
			if errs := schema.Validate(doc); len(errs) > 0 {
				t.Fatalf("domain fails embedded domain-spec-v0.1 schema: %s", errs[0])
			}
		})
	}
	if count != 6 {
		t.Fatalf("expected six shipped domains, found %d", count)
	}
}
