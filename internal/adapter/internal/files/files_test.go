package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestLoadReadsAndValidatesAnAdapterFile(t *testing.T) {
	t.Parallel()
	a, err := Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != "native-jsonl" || len(a.Raw) == 0 {
		t.Fatalf("adapter = %q with %d raw bytes", a.ID, len(a.Raw))
	}
}

func TestLoadNamesTheMissingFile(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent.adapter.json")
	if _, err := Load(missing); err == nil || !strings.Contains(err.Error(), "adapter: read "+missing) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadDirSkipsNonJSONFilesAndNamesTheAdapterItCannotLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	native, err := os.ReadFile(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{"native.adapter.json": native, "notes.txt": []byte("not an adapter")} {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	adapters, err := LoadDir(dir)
	if err != nil || len(adapters) != 1 || adapters["native-jsonl"] == nil {
		t.Fatalf("adapters = %v, %v; want only the JSON file loaded", adapters, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDir(dir); err == nil || !strings.HasPrefix(err.Error(), "streamsim: ") {
		t.Fatalf("a broken adapter: err = %v", err)
	}
}
