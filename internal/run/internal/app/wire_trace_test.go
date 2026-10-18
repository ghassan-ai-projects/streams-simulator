package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAWireTraceWritesEveryRecordItIsGivenAsAnObserverWireSessionRecord(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	trace, err := OpenWireTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	trace.Wire("gateway", map[string]any{"type": "DDATA", "seq": 3})
	trace.Wire("node", map[string]any{"type": "DDATA", "seq": 2})
	if err := trace.Flush(); err != nil {
		t.Fatal(err)
	}
	flushed, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile))
	if err != nil || strings.Count(string(flushed), `"type":"wire"`) != 2 {
		t.Fatalf("journal after Flush = %q, %v, want both records on disk", flushed, err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	closed, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile))
	if incomplete, ended := journal.ClosedIncomplete(closed); err != nil || incomplete || !ended {
		t.Fatalf("journal after Close = %q, %v, want it closed and complete", closed, err)
	}
	if !strings.Contains(string(closed), `"direction":"gateway"`) || !strings.Contains(string(closed), `"stream":"observer"`) {
		t.Fatalf("journal = %q, want the direction and the observer stream on each record", closed)
	}
}

func TestAWireTraceCannotBeOpenedUnderAFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "plain", nil)
	_, err := OpenWireTrace(filepath.Join(dir, "plain", "run"))
	var pathError *fs.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("OpenWireTrace = %v, want the file system's refusal", err)
	}
}
