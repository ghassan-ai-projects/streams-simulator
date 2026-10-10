package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := root.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAllReadsOnlyTopLevelJSONInSortedOrder(t *testing.T) {
	t.Parallel()
	example, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeFile(t, dir, "b.domain.json", string(example))
	writeFile(t, dir, "a.domain.json", string(example))
	writeFile(t, dir, "notes.txt", "not a domain")
	if err := os.Mkdir(filepath.Join(dir, "nested.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	specs, err := LoadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("loaded %d specs, want 2", len(specs))
	}
}

func TestLoadAllFailsTheWholeLoadOnOneBadDomain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "broken.json", `{"id":`)
	_, err := LoadAll(dir)
	if err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Fatalf("err = %v, want a failure naming broken.json", err)
	}
}

func TestLoadAndLoadAllNameTheMissingPath(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent")
	if _, err := Load(missing); err == nil || !strings.Contains(err.Error(), "domain: read "+missing) {
		t.Fatalf("Load err = %v", err)
	}
	if _, err := LoadAll(missing); err == nil || !strings.Contains(err.Error(), "domain: list "+missing) {
		t.Fatalf("LoadAll err = %v", err)
	}
}
