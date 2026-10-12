package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
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

func TestLoadAllReadsOnlyTopLevelJSON(t *testing.T) {
	t.Parallel()
	example, err := os.ReadFile(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeFile(t, dir, "a.domain.json", string(example))
	writeFile(t, dir, "notes.txt", "not a domain")
	if err := os.Mkdir(filepath.Join(dir, "nested.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	specs, err := LoadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 {
		t.Fatalf("loaded %d specs, want 1 (txt and nested dir ignored)", len(specs))
	}
}

func TestLoadAllVisitsFilesInSortedOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "b.json", `{"id":`)
	writeFile(t, dir, "a.json", `{"id":`)
	_, err := LoadAll(dir)
	if err == nil || !strings.Contains(err.Error(), "a.json") || strings.Contains(err.Error(), "b.json") {
		t.Fatalf("err = %v, want the failure to name a.json only", err)
	}
}

func TestLoadAllFailsTheWholeLoadOnOneBadDomain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "broken.json", `{"id":`)
	_, err := LoadAll(dir)
	if err == nil || !strings.Contains(err.Error(), "broken.json") || !strings.HasPrefix(err.Error(), "streamsim: ") {
		t.Fatalf("err = %v, want a streamsim-prefixed failure naming broken.json", err)
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

func TestFormatRewritesAFileInPlaceAndKeepsItsMode(t *testing.T) {
	t.Parallel()
	example, err := os.ReadFile(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeFile(t, dir, "a.domain.json", strings.ReplaceAll(string(example), "\n", "\n\n"))
	path := filepath.Join(dir, "a.domain.json")
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	result, err := Format(path)
	rewritten, readErr := os.ReadFile(path)
	if err != nil || readErr != nil || !result.Changed || string(rewritten) != string(example) {
		t.Fatalf("Format = %+v (%v, %v), want the canonical example bytes", result, err, readErr)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o400 {
		t.Fatalf("mode = %v (%v), want 0400 kept", info, err)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".format-*")); len(left) != 0 {
		t.Fatalf("staging files left behind: %v", left)
	}
	again, err := Format(path)
	if err != nil || again.Changed || again.Digest != result.Digest {
		t.Fatalf("second Format = %+v (%v), want no change", again, err)
	}
}

func TestFormatLeavesAnInvalidFileAlone(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "bad.json", `{"id":"X"}`)
	if _, err := Format(filepath.Join(dir, "bad.json")); err == nil || !strings.Contains(err.Error(), "domain-spec-v0.2") {
		t.Fatalf("err = %v, want a domain-spec-v0.2 refusal", err)
	}
	kept, err := os.ReadFile(filepath.Join(dir, "bad.json"))
	if err != nil || string(kept) != `{"id":"X"}` {
		t.Fatalf("file = %q (%v), want it untouched", kept, err)
	}
}
