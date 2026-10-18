package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestAWireTraceFromTheFacadeJournalsWhatItIsGiven(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	trace, err := run.NewWireTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	trace.Wire("node", map[string]any{"type": "NBIRTH", "seq": 0})
	if err := trace.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile))
	if err != nil || !strings.Contains(string(raw), `"direction":"node"`) {
		t.Fatalf("journal = %q, %v, want the wire record of the node direction", raw, err)
	}
}
