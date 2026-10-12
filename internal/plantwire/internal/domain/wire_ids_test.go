package domain

import (
	"encoding/json"
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

func TestArgumentSchemasKeepOnlyValidationKeywordsAtEveryDepth(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"description": "x", "title": "t", "$comment": "c", "examples": []any{1}, "default": 3, "x-note": "n",
		"type": "object", "required": []any{"description"},
		"properties": map[string]any{"pct": map[string]any{"type": "number", "description": "y", "maximum": 100},
			"description": map[string]any{"type": "string", "title": "named description"}},
		"anyOf": []any{map[string]any{"description": "z", "type": "object"}}, "items": map[string]any{"enum": []any{"a"}, "title": "i"}}
	got, _ := json.Marshal(validationOnly(schema))
	want := `{"anyOf":[{"type":"object"}],"items":{"enum":["a"]},"properties":{"description":{"type":"string"},` +
		`"pct":{"maximum":100,"type":"number"}},"required":["description"],"type":"object"}`
	if string(got) != want {
		t.Fatalf("validation-only schema\n got %s\nwant %s", got, want)
	}
}

func TestNegotiationRespectsTheHostsRange(t *testing.T) {
	t.Parallel()
	if _, err := Negotiate(Message{Online: true, ProtocolVersion: 1, ProtocolMin: 2, ProtocolMax: 3}, 3); !errors.Is(err, ErrRefused) {
		t.Fatalf("negotiate = %v, want a refusal of a version below the host's minimum", err)
	}
}
