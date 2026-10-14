package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	openedLine = `{"type":"run.opened","mode":"evaluate"}` + "\n"
	closedLine = `{"type":"run.closed"}` + "\n"
	sealedFile = "runs/w-1/sealed/journal.jsonl"
)

func TestAStaleRunArtifactDoesNotUnlockTheSealedFilesOfAnOpenRun(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-1/journal.jsonl", openedLine)
	h.write(t, "runs/w-1/run.json", "{}")
	h.write(t, "runs/w-1/ledger.jsonl", "truth\n")
	h.write(t, sealedFile, "truth\n")
	for _, name := range []string{"run.json", "ledger.jsonl", "sealed/journal.jsonl"} {
		if _, err := h.bridge.ReadFile("runs/w-1/" + name); !errors.Is(err, domain.ErrSealed) {
			t.Errorf("%s of a run whose journal is not closed: %v", name, err)
		}
	}
}

func TestARunIsEndedOnlyOnceItsJournalClosedAndItsWorldStoppedRunning(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	h.prepare = func(d *fakeDirector) {
		d.results["sim.world.create"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1"}`)}
	}
	if _, err := h.bridge.RunAction(ctx, "world.create", map[string]any{"domain": "plant"}); err != nil {
		t.Fatal(err)
	}
	h.write(t, "runs/w-1/journal.jsonl", openedLine+closedLine)
	h.write(t, sealedFile, "truth\n")
	if _, err := h.bridge.ReadFile(sealedFile); !errors.Is(err, domain.ErrSealed) {
		t.Fatalf("a closed journal of a world the bridge still runs: %v", err)
	}
	if _, err := h.bridge.RunAction(ctx, "run.end", map[string]any{"world_id": "w-1"}); err != nil {
		t.Fatal(err)
	}
	if document, err := h.bridge.ReadFile(sealedFile); err != nil || document.Content != "truth\n" {
		t.Fatalf("after run.end the sealed journal is readable: %+v, %v", document, err)
	}
}

func TestFollowingAndOpeningARunTakeOneRunIdAndNeverASealedPath(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-2/journal.jsonl", openedLine)
	h.write(t, "runs/w-2/sealed/journal.jsonl", `{"type":"sealed.secret"}`+"\n")
	for _, id := range []string{"w-2/sealed", "w-2/../w-2", "..", "", "w 2"} {
		if err := h.bridge.OpenRun(id); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("OpenRun(%q) = %v, want ErrBadPath", id, err)
		}
		emitted := 0
		err := h.bridge.Follow(context.Background(), id, -1, func(domain.Record) error { emitted++; return nil })
		if !errors.Is(err, domain.ErrBadPath) || emitted != 0 {
			t.Errorf("Follow(%q) = %v after %d records, want ErrBadPath and none", id, err, emitted)
		}
	}
	if err := h.bridge.OpenRun("w-2"); err != nil {
		t.Fatalf("OpenRun of a run: %v", err)
	}
	if err := h.bridge.OpenRun("w-9"); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("OpenRun of a run that does not exist: %v", err)
	}
}

func TestResumingAfterTheClosingRecordStillTellsTheStreamItClosed(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-1/journal.jsonl", openedLine+`{"type":"observation"}`+"\n"+closedLine)
	var got []domain.Record
	err := h.bridge.Follow(context.Background(), "w-1", 2, func(record domain.Record) error {
		got = append(got, record)
		return nil
	})
	if err != nil || len(got) != 1 || !got[0].Closes || !got[0].Seen || got[0].Index != 2 {
		t.Fatalf("Follow after the closing record: %+v, %v", got, err)
	}
}

func TestARunReadThroughASymlinkOutOfTheWorkspaceIsABadPath(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(h.domainsDir(), "link.domain.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := h.bridge.ReadFile(inDomains("link.domain.json")); !errors.Is(err, domain.ErrBadPath) {
		t.Fatalf("a symlink out of the workspace: %v", err)
	}
}
