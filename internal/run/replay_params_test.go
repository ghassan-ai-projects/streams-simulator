package run

// Regression: command-log params round-trip through the artifact as
// json.Number, so validation that accepts only float64/int64 breaks replay.
// The soak caught this at 1M records; this test locks it small.

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestReplayRoundTripsCommandParams(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 0.5}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("site-a/pond-1", "aerator_failure", start+3600*1e9, map[string]any{"severity": 2.0}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
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
		t.Fatalf("replay must accept artifact-decoded params: %v", err)
	}
	if !res.Matches {
		t.Fatalf("replay diverged: %s vs %s", res.GotDigest, res.WantDigest)
	}
}
