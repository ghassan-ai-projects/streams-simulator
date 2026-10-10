package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// TestReplayDetectsTampering: a corrupted expected digest must fail verify.
func TestReplayDetectsTampering(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}
	art := buildArtifact(t, cfg)
	art.ExpectedTraceDigest = "sha256:" + strings.Repeat("0", 64)
	res, err := ReplayArtifact(context.Background(), art, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Matches {
		t.Fatal("tampered digest must not match")
	}
}

func TestReplayRejectsInputDigestMismatch(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	art := buildArtifact(t, Config{
		Domain: spec, Adapter: a, Seed: 6, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	})
	bad := *art
	bad.Domain = art.Domain
	bad.Domain.Digest = "sha256:" + strings.Repeat("0", 64)
	if _, err := ReplayArtifact(context.Background(), &bad, spec, a, ""); err == nil {
		t.Fatal("replay must reject a changed domain digest before execution")
	}
}

// TestArtifactRoundTrip validates the artifact against its schema and the
// command log is a faithful record of the run.
func TestArtifactRoundTrip(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{
		Domain: spec, Adapter: a, Seed: 9, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}
	dir := t.TempDir()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("site-a/pond-1", "do_probe_fouling", 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), cfg.StartTimeNS+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	// The artifact validates against the committed schema.
	raw, err := json.Marshal(art)
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		t.Fatalf("artifact fails its own schema: %v", err)
	}
	// Replay from the on-disk artifact reproduces the run.
	loaded, err := LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReplayArtifact(context.Background(), loaded, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Matches {
		t.Fatalf("replay of on-disk artifact failed: %s vs %s", res.GotDigest, res.WantDigest)
	}
	// Files written: trace, ledger, history, run.json.
	for _, name := range []string{"trace.jsonl", "ledger.jsonl", "world_state_history.jsonl", "run.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing artifact file %s: %v", name, err)
		}
	}
}

// TestEnvInjectRejectsUndefinedParams: environment faults declare no params
// yet, so any params are rejected rather than recorded and ignored.
func TestEnvInjectRejectsUndefinedParams(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	r.ConfigureEnvTarget("consumer-1", true)
	if _, err := r.EnvInject("consumer-1", "pause", map[string]any{"duration_s": 30}, start); err == nil {
		t.Fatal("env.inject params must be rejected (none declared)")
	}
	if _, err := r.EnvInject("consumer-1", "pause", nil, start); err != nil {
		t.Fatalf("env.inject without params rejected: %v", err)
	}
}

func TestReplayDivergenceSaysWhatIsAndIsNotKnown(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}
	wrongDigest := "sha256:" + strings.Repeat("0", 64)

	sameCount := buildArtifact(t, cfg)
	sameCount.ExpectedTraceDigest = wrongDigest
	res, err := ReplayArtifact(context.Background(), sameCount, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Matches || res.FirstDivergence != nil || !strings.Contains(res.Detail, "first differing record is not known") {
		t.Fatalf("same count, different digest: %+v", res)
	}

	longer := buildArtifact(t, cfg)
	longer.ExpectedTraceDigest = wrongDigest
	longer.Counts.Emitted += 5
	res, err = ReplayArtifact(context.Background(), longer, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.FirstDivergence == nil || int64(*res.FirstDivergence) != longer.Counts.Emitted-5 {
		t.Fatalf("a replay shorter than the artifact diverges at the replayed length: %+v", res)
	}
	if !strings.Contains(res.Detail, "recorded") {
		t.Fatalf("detail = %q", res.Detail)
	}
}
