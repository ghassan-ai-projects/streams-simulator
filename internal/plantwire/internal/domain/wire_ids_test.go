package domain

import (
	"errors"
	"math"
	"testing"
)

func TestWireIDsAreSafeDistinctAndKeepSafeIdsAsTheyAre(t *testing.T) {
	t.Parallel()
	for id, want := range map[string]string{"pump-3": "pump-3", "site-a/pond-1": "site-a:pond-1", "/root": "d:root", "a b": "a_b", "café": "caf_"} {
		if got := WireID(id); got != want || !wireID.MatchString(got) {
			t.Errorf("WireID(%q) = %q, want %q", id, got, want)
		}
	}
	taken := map[string]bool{"a:b": true, "a:b.2": true}
	if got := Distinct("a:b", func(c string) bool { return taken[c] }); got != "a:b.3" {
		t.Fatalf("Distinct = %q", got)
	}
	if !Finite(1.5) || !Finite("text") || Finite(math.NaN()) || Finite(math.Inf(1)) {
		t.Fatal("Finite")
	}
}

func TestArgumentSchemasLoseTheirDescriptionsAtEveryDepth(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"description": "x", "properties": map[string]any{"pct": map[string]any{"type": "number", "description": "y"}},
		"anyOf": []any{map[string]any{"description": "z", "type": "object"}}}
	got := withoutDescriptions(schema).(map[string]any)
	if got["description"] != nil || got["properties"].(map[string]any)["pct"].(map[string]any)["description"] != nil ||
		got["anyOf"].([]any)[0].(map[string]any)["description"] != nil || got["properties"].(map[string]any)["pct"].(map[string]any)["type"] != "number" {
		t.Fatalf("stripped %v", got)
	}
}

func TestNegotiationRespectsTheHostsRange(t *testing.T) {
	t.Parallel()
	if _, err := Negotiate(map[string]any{"online": true, "protocol_version": 1, "protocol_min": 2, "protocol_max": 3}, 3); !errors.Is(err, ErrRefused) {
		t.Fatalf("negotiate = %v, want a refusal of a version below the host's minimum", err)
	}
}
