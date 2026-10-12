package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestWorldWritesBothJournalStreamsInItsRunDirectory(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	world := d.RunOf(worldID)
	fault := world.Domain().Spec.Kinds[0].Faults[0].ID
	if _, err := d.InjectFault(worldID, world.World.EntityIDs()[0], fault, world.World.Clock()+60e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(t.Context(), worldID, world.World.Clock()+600e9, false); err != nil {
		t.Fatal(err)
	}
	ended, err := d.finalizeRun(d.World(worldID), worldID)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(ended["run_artifact_path"].(string))
	observer, sealed := readJournalFile(t, dir, journal.ObserverFile), readJournalFile(t, dir, journal.DirectorFile)
	if bytes.Contains(observer, []byte(fault)) || !bytes.Contains(sealed, []byte(fault)) {
		t.Fatalf("the fault %q belongs to the director stream only", fault)
	}
	if bytes.Contains(observer, []byte(`"seed"`)) || !bytes.Contains(sealed, []byte(`"seed"`)) {
		t.Fatal("the seed belongs to the director stream only")
	}
}

func readJournalFile(t *testing.T, dir, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestExploreWorldIsUnblindedFromCreationAndRefusesScoring(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(WorldRequest{Domain: "aquaculture-pond", Seed: 3, Adapter: "native-jsonl", Explore: true})
	if err != nil {
		t.Fatal(err)
	}
	worldID, runID := created["world_id"].(string), created["run_id"].(string)
	if !d.RunOf(worldID).UnblindedStamp() {
		t.Fatal("an explore world's run is unblinded from creation")
	}
	var refusal *rules.ToolError
	if _, err := d.Score(runID); !errors.As(err, &refusal) || refusal.Code != rules.CodeRunUnblinded {
		t.Fatalf("Score of an explore run = %v, want %s", err, rules.CodeRunUnblinded)
	}
	ended, err := d.finalizeRun(d.World(worldID), worldID)
	if err != nil {
		t.Fatal(err)
	}
	art, err := run.LoadArtifact(ended["run_artifact_path"].(string))
	if err != nil || art.Mode != journal.ModeExplore || !art.Unblinded {
		t.Fatalf("artifact = %+v, %v", art, err)
	}
}

func TestRunOfNamesTheRunBehindAWorldAndNilForAnUnknownOne(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	if d.RunOf(worldID) != d.World(worldID).Run || d.RunOf("w-404") != nil {
		t.Fatal("RunOf must return the world's run, nil for an unknown world")
	}
}
