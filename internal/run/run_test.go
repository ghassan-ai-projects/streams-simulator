package run_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func pondRun(t *testing.T, sinkDir string) *run.Run {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(filepath.Join("..", "..", "adapters", "native-jsonl.adapter.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS}
	if sinkDir != "" {
		cfg.SinkName = model.SinkFile
		cfg.SinkTarget = filepath.Join(sinkDir, "trace.jsonl")
	}
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunExposesItsIdentityAndRecordsCommands(t *testing.T) {
	t.Parallel()
	r := pondRun(t, "")
	if r.ID == "" || r.World == nil || r.Config.Seed != 4 || r.Domain() == nil {
		t.Fatalf("identity = %q %v %d", r.ID, r.World, r.Config.Seed)
	}
	pond := r.World.EntityIDs()[0]
	start := model.DefaultStartTimeNS
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 0.5}, start, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(pond, "aerator_failure", start+60e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+600e9, false); err != nil {
		t.Fatal(err)
	}
	if len(r.Ledger()) == 0 || len(r.History()) == 0 || len(r.AppliedPerturbations()) != 1 {
		t.Fatalf("ledger %d history %d perturbations %v", len(r.Ledger()), len(r.History()), r.AppliedPerturbations())
	}
	if !r.Reproducible() || r.Digest() == "" || r.UnblindedStamp() {
		t.Fatal("a stepped inproc run is reproducible, blind and has a world digest")
	}
}

func TestEndPublishesAnArtifactThatReplaysToTheSameTrace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	r := pondRun(t, dir)
	if _, err := r.Advance(context.Background(), model.DefaultStartTimeNS+300e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := run.LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil || loaded.ExpectedTraceDigest != art.ExpectedTraceDigest {
		t.Fatalf("loaded digest %q, ended %q (%v)", loaded.ExpectedTraceDigest, art.ExpectedTraceDigest, err)
	}
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := run.ReplayArtifact(context.Background(), loaded, spec, nil, "")
	if err != nil || !result.Matches {
		t.Fatalf("replay = %+v (%v)", result, err)
	}
}

func TestAnInvalidVerdictIsRefusedAndNeverKept(t *testing.T) {
	t.Parallel()
	r := pondRun(t, "")
	if r.Verdict() != nil {
		t.Fatal("no verdict before one is submitted")
	}
	if err := r.SubmitVerdict(&model.Verdict{}); err == nil {
		t.Fatal("an empty verdict must be refused")
	}
	if r.Verdict() != nil {
		t.Fatal("a refused verdict must not be kept")
	}
	r.Unblind()
	if !r.UnblindedStamp() {
		t.Fatal("unblinding must stamp the run")
	}
}

func TestCommandsAreRecordedAndTheTestHooksReachTheRun(t *testing.T) {
	t.Parallel()
	r := pondRun(t, "")
	start := model.DefaultStartTimeNS
	pond := r.World.EntityIDs()[0]
	var seen int
	r.SetEvidenceRecorder(func(model.SimEvent) { seen++ })
	r.SetQuiesceParkedHook(func() {})
	r.SetFailureMode("silent_no_effect")
	id, err := r.ApplyPerturb("duplicate_burst", map[string]any{"rate": 0.2}, start, 0)
	if err != nil {
		t.Fatal(err)
	}
	faultID, err := r.InjectFault(pond, "aerator_failure", start+60e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond}, start+30e9); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+600e9, false); err != nil {
		t.Fatal(err)
	}
	r.ReportQuiesced(start + 600e9)
	if err := r.ClearPerturb(id); err != nil {
		t.Fatal(err)
	}
	if err := r.ClearFault(faultID, start+700e9); err != nil {
		t.Fatal(err)
	}
	if err := r.RetireEntity(pond, "test", start+800e9); err != nil {
		t.Fatal(err)
	}
	if _, err := r.EnvInject("none", "f", nil, start+900e9); err == nil {
		t.Fatal("environment injection is not enabled for a plain run")
	}
	if seen == 0 {
		t.Fatal("the evidence recorder must see delivered events")
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	if len(r.Trace()) == 0 {
		t.Fatal("the closed run must expose its trace")
	}
}
