package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestDomainFormatRewritesCanonicallyAndIsIdempotent(t *testing.T) {
	t.Parallel()
	shipped, err := os.ReadFile(testsupport.Domain(pondDomain))
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(shipped))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	compact, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	path := writeFile(t, pondDomain+".domain.json", string(compact))
	first := invoke(t, "domain", "format", path).mustSucceed(t).document(t, "domain_format")
	second := invoke(t, "domain", "format", path).mustSucceed(t).document(t, "domain_format")
	rewritten, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(rewritten, shipped) {
		t.Fatalf("formatted file differs from the shipped canonical form (%v)", err)
	}
	if first["changed"] != true || second["changed"] != false || first["digest"] != second["digest"] {
		t.Fatalf("first %v, second %v: want a change once, then none, with one digest", first, second)
	}
}

func TestDomainFormatRefusesAnInvalidDomainAndLeavesTheFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "bad.domain.json")
	if err := os.WriteFile(path, []byte(`{"id":"X"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	failure := invoke(t, "domain", "format", path).failure(t)
	kept, err := os.ReadFile(path)
	if failure["code"] != "domain_invalid" || err != nil || string(kept) != `{"id":"X"}` {
		t.Fatalf("failure %v, file %q (%v): want domain_invalid and the file untouched", failure, kept, err)
	}
}
