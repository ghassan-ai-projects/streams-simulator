package files

import (
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
