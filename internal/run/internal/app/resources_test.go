package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// openFileDescriptors counts this process's open descriptors.
func openFileDescriptors(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("no /dev/fd on this platform: %v", err)
	}
	return len(entries)
}

// A failed start must hand back what it opened. It counts this process's
// descriptors, so it runs while no parallel test does.
//
//nolint:paralleltest // descriptor counts are process-wide
func TestFailedNewLeaksNoFileDescriptor(t *testing.T) {
	spec, a := testBase(t)
	dir := t.TempDir()
	before := openFileDescriptors(t)
	for i := range 10 {
		cfg := Config{Domain: spec, Adapter: a, SinkName: "no-such-sink", LedgerPath: filepath.Join(dir, "ledger.jsonl")}
		if _, err := New(t.Context(), cfg); err == nil || !strings.Contains(err.Error(), "unsupported sink") {
			t.Fatalf("attempt %d: err = %v", i, err)
		}
	}
	if after := openFileDescriptors(t); after > before {
		t.Fatalf("ten failed starts leaked %d file descriptors", after-before)
	}
}

func TestEndKeepsTheArtifactWhenItsEvidenceCannotBePublished(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc})
	if err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	art, err := r.End(filepath.Join(blocker, "out"))
	if err == nil || !strings.HasPrefix(err.Error(), "End: ") {
		t.Fatalf("publishing under a regular file: err = %v", err)
	}
	if art == nil || art.ExpectedTraceDigest == "" || !r.Finished() {
		t.Fatalf("the finished run's artifact must survive a publication failure: art=%v finished=%v", art, r.Finished())
	}
	if _, err := r.End(""); err == nil || !strings.Contains(err.Error(), "already finished") {
		t.Fatalf("a second End: %v", err)
	}
}
