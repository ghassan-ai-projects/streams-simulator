package domain

// Every shipped domain must load, validate and carry a stable digest. This
// is the catalog's integrity check: a domain that silently stops loading
// would make coverage reports lie.

import (
	"path/filepath"
	"testing"
)

func TestAllShippedDomainsLoad(t *testing.T) {
	dir := filepath.Join("..", "..", "domains")
	list, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("a shipped domain failed to load: %v", err)
	}
	if len(list) < 6 {
		t.Fatalf("expected at least the six catalog domains, got %d", len(list))
	}
	seen := map[string]bool{}
	for _, c := range list {
		if seen[c.Spec.ID] {
			t.Fatalf("duplicate domain id %q", c.Spec.ID)
		}
		seen[c.Spec.ID] = true
		if !c.HasProfile("nominal") || !c.HasProfile("correlated_cascade") || !c.HasProfile("sensor_pathology") {
			t.Fatalf("domain %s must declare nominal/correlated_cascade/sensor_pathology profiles", c.Spec.ID)
		}
		hasNegative := false
		for i := range c.Spec.Faults {
			if c.Spec.Faults[i].IsNegativeClass {
				hasNegative = true
			}
		}
		if !hasNegative {
			t.Fatalf("domain %s must declare a negative-class fault", c.Spec.ID)
		}
	}
	// The docs example and the installed copy must agree byte-for-byte.
	ex := filepath.Join("..", "..", "docs", "examples", "aquaculture-pond.domain.json")
	installed := filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json")
	a, err := Load(ex)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(installed)
	if err != nil {
		t.Fatal(err)
	}
	if a.Digest != b.Digest {
		t.Fatalf("installed aquaculture-pond drifted from docs/examples: %s vs %s", a.Digest, b.Digest)
	}
}
