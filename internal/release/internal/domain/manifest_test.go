package domain

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestBodyRecordsIdentityDigestsAndTheClosedSuiteSettings(t *testing.T) {
	t.Parallel()
	identity := Identity{Version: "v1", Commit: "abc", Author: "A <a@x>", Reviewer: "R <r@x>", CreatedAt: "2026-10-10T12:00:00Z"}
	body := Body(identity, map[string]any{"d.json": "sha256:1"}, map[string]any{"a.json": "sha256:2"})
	raw, err := Canonical(body)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	sim, _ := decoded["sim"].(map[string]any)
	if decoded["schema_version"] != SchemaVersion || sim["version"] != "v1" || sim["commit"] != "abc" || !strings.HasPrefix(sim["go"].(string), "go") {
		t.Fatalf("body = %v", decoded)
	}
	if decoded["author"] != identity.Author || decoded["reviewer"] != identity.Reviewer || decoded["created_at"] != identity.CreatedAt {
		t.Fatalf("identity = %v", decoded)
	}
}

func TestDocumentPlacesTheCanonicalBodyBesideItsSignature(t *testing.T) {
	t.Parallel()
	raw, err := Document([]byte(`{"a":1}`), "sig")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Manifest  map[string]any `json:"manifest"`
		Signature string         `json:"signature"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil || doc.Manifest["a"] != float64(1) || doc.Signature != "sig" {
		t.Fatalf("document = %s (%v)", raw, err)
	}
}

func TestFileNamesSkipHiddenFilesAndSort(t *testing.T) {
	t.Parallel()
	got := FileNames([]string{"b.json", ".DS_Store", "a.json"})
	if !slices.Equal(got, []string{"a.json", "b.json"}) {
		t.Fatalf("names = %v", got)
	}
}
