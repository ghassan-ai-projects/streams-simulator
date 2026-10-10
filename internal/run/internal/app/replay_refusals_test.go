package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// A refused invocation is part of the command log so that replay reproduces
// it; the refusal the original run got must not abort the replay.
func TestReplayReproducesARefusedEffectorInvocation(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 7, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS,
	})
	if err != nil {
		t.Fatal(err)
	}
	pond := r.World.EntityIDs()[0]
	now := r.World.Clock()
	args := map[string]any{"pond_id": pond, "level": 1.0}
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-1", args, now); err != nil {
		t.Fatal(err)
	}
	reused := map[string]any{"pond_id": pond, "level": 0.5}
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-1", reused, now); err == nil {
		t.Fatal("a command_id reused for a different request must be refused")
	}
	dir := t.TempDir()
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReplayArtifact(context.Background(), loaded, spec, a, "")
	if err != nil {
		t.Fatalf("replay must reproduce the refusal, not fail on it: %v", err)
	}
	if !res.Matches {
		t.Fatalf("replay diverged: %s vs %s", res.GotDigest, res.WantDigest)
	}
}
