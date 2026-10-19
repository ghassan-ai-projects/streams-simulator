package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func (h *harness) blockFolder(t *testing.T, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(h.dir, filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
	h.write(t, rel, "not a folder")
}

func TestAWorkspaceFolderThatCannotBeListedFailsTheListingNamingIt(t *testing.T) {
	t.Parallel()
	for _, folder := range []string{"domains", "kinds", "components/equipment", "runs"} {
		t.Run(folder, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.blockFolder(t, folder)
			if _, err := h.bridge.Workspace(); err == nil || !strings.Contains(err.Error(), "canvasbridge: list "+folder) {
				t.Fatalf("Workspace() = %v, want an error naming %s", err, folder)
			}
		})
	}
}

func TestARunWhoseJournalCannotBeReadIsListedWithoutASummary(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if err := os.MkdirAll(filepath.Join(h.dir, "runs", "w-1", "journal.jsonl"), 0o750); err != nil {
		t.Fatal(err)
	}
	listing, err := h.bridge.Workspace()
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("Workspace() runs = %+v, %v, want the one run listed", listing.Runs, err)
	}
	if run := listing.Runs[0]; run.WorldID != "w-1" || run.Ended || run.Plant != "" || run.RecordedAt == "" {
		t.Fatalf("run = %+v, want w-1 with no summary and its recorded time", run)
	}
}

func TestAGuardedWriteWhoseTargetCannotBeReadIsRefusedBeforeAnyCommand(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if err := os.MkdirAll(filepath.Join(h.dir, "scenarios", "s.scenario.json"), 0o750); err != nil {
		t.Fatal(err)
	}
	_, err := h.bridge.RunAction(context.Background(), "scenario.export", map[string]any{"run": "w-1", "out": "scenarios/s.scenario.json"})
	if err == nil || !strings.Contains(err.Error(), "scenarios/s.scenario.json") || len(h.argv) != 0 {
		t.Fatalf("exporting over an unreadable target = %v after %d commands, want the read error and no command", err, len(h.argv))
	}
}

func TestACommandThatCannotRunFailsTheCallWithItsError(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.bridge.edges.RunCommand = func(context.Context, []string, []byte) (process.CommandResult, error) {
		return process.CommandResult{}, errors.New("no such executable")
	}
	if _, err := h.bridge.RunAction(context.Background(), "run.verify", map[string]any{"run": "w-1"}); err == nil || err.Error() != "no such executable" {
		t.Fatalf("run.verify with no executable = %v, want the launch error", err)
	}
}

func TestADirectorThatCannotStartFailsTheToolCallWithItsError(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.bridge.edges.StartDirector = func(context.Context) (Director, error) { return nil, errors.New("no director binary") }
	if _, err := h.bridge.RunAction(context.Background(), "catalog.list", nil); err == nil || err.Error() != "no director binary" {
		t.Fatalf("catalog.list without a director = %v, want the start error", err)
	}
}
