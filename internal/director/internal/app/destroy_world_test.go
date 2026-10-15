package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// A run whose End failed is finished all the same. Destroying its world must
// leave the world destroyable: the retry may not fail with "already finished".
func TestDestroyWorldAfterAFailedEndCanBeRetried(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	// Replace the world's artifact directory with a file: End cannot write.
	dir := filepath.Join(d.OutDir, worldID)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := d.DestroyWorld(worldID)
	var toolErr *rules.ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != rules.CodeDomainInvalid {
		t.Fatalf("err = %v, want a %s tool error for the failed End", err, rules.CodeDomainInvalid)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatalf("the retry must destroy the world: %v", err)
	}
	if d.World(worldID) != nil {
		t.Fatal("the world is still registered")
	}
}

// A world destroyed before run.end keeps its evidence as an incomplete run with the reason.
func TestDestroyingAWorldBeforeItsRunEndsKeepsAnIncompleteRun(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	world := d.RunOf(worldID)
	_, err := d.InjectFault(worldID, world.World.EntityIDs()[0], "no-such-fault", world.World.Clock()+60e9, nil)
	if err == nil || !strings.Contains(err.Error(), "no-such-fault") {
		t.Fatalf("an unknown fault must be refused by name: %v", err)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(d.OutDir, worldID)
	art, err := run.LoadArtifact(filepath.Join(dir, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	if !art.Incomplete || art.Reproducible || !strings.Contains(art.Error, "destroyed before its run ended") {
		t.Fatalf("run.json = incomplete %v, reproducible %v, error %q; want an incomplete run with the reason", art.Incomplete, art.Reproducible, art.Error)
	}
	if incomplete, closed := journal.ClosedIncomplete(readJournalFile(t, dir, journal.ObserverFile)); !closed || !incomplete {
		t.Fatalf("the observer journal closed incomplete=%v closed=%v, want incomplete", incomplete, closed)
	}
}

func TestDestroyingAWorldAfterItsRunEndedKeepsACompleteRun(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	if _, err := d.EndRun(worldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatal(err)
	}
	art, err := run.LoadArtifact(filepath.Join(d.OutDir, worldID, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	if art.Incomplete || art.Error != "" {
		t.Fatalf("a run that ended normally was recorded incomplete=%v error=%q", art.Incomplete, art.Error)
	}
}
