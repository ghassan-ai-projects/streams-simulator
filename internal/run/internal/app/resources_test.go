package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// A failed start must hand back what it opened: the ledger file it opened is
// closed (closing it again reports os.ErrClosed).
func TestFailedNewClosesTheLedgerItOpened(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{Domain: spec, Adapter: a, SinkName: "no-such-sink", LedgerPath: filepath.Join(t.TempDir(), "ledger.jsonl")}
	if _, err := New(t.Context(), cfg); err == nil || !strings.Contains(err.Error(), "unsupported sink") {
		t.Fatalf("New = %v, want the unsupported sink", err)
	}
	r, err := newRun(t.Context(), defaultConfig(cfg))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.beginReleasing(); err == nil || !strings.Contains(err.Error(), "unsupported sink") {
		t.Fatalf("beginReleasing = %v, want the unsupported sink", err)
	}
	if r.durableLedger == nil {
		t.Fatal("the ledger was never opened, so the test proves nothing")
	}
	if err := r.durableLedger.Finish(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("closing the ledger of a failed start again = %v, want os.ErrClosed (it was left open)", err)
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

// Every delivery gets a ledger row, also after the sink has failed: the
// duplicate the perturbation made of an event is accounted for even though
// the event itself could not be written.
func TestSinkFailureStillAccountsForEveryDelivery(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkHTTPPush, SinkTarget: "http://127.0.0.1:1/unreachable"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("duplicate_burst", map[string]any{"rate": 1.0}, r.World.Clock(), 0); err != nil {
		t.Fatal(err)
	}
	_, _ = r.Advance(t.Context(), r.World.Clock()+120*1e9, false)
	emitted := int(r.World.EmittedCount())
	ledger := r.Ledger()
	if emitted == 0 || len(ledger) < 2*emitted {
		t.Fatalf("%d emitted events but %d ledger rows: a duplicate went unaccounted for", emitted, len(ledger))
	}
	for _, row := range ledger {
		if row.Delivered || row.DeliveryReason != model.DeliverySinkError {
			t.Fatalf("with a dead sink every row is a sink error, got %+v", row)
		}
	}
}
