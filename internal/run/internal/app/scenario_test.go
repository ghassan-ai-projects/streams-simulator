package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// drivenArtifact runs the example pond by hand (a fault, a perturbation, two
// effector calls and two advances) and returns its artifact.
func drivenArtifact(t *testing.T, cfg Config) *model.RunArtifact {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	pond, start := r.World.EntityIDs()[0], cfg.StartTimeNS
	call := map[string]any{"pond_id": pond, "level": 1.0}
	steps := []error{
		firstError(r.InjectFault(pond, "aerator_failure", start+3600e9, nil)),
		firstError(r.ApplyPerturb("drop", nil, start+60e9, start+120e9)),
		firstError(r.InvokeEffector("start_aerator", pond, "c-1", call, start)),
		firstError(r.Advance(context.Background(), start+2*3600e9, false)),
		firstError(r.InvokeEffector("start_aerator", pond, "c-2", call, start+2*3600e9)),
		firstError(r.Advance(context.Background(), start+4*3600e9, false)),
	}
	if err := errors.Join(steps...); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

func firstError[T any](_ T, err error) error { return err }

func TestAScenarioExportedFromARunReproducesIt(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{Domain: spec, Adapter: a, Seed: 8, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true}
	original := drivenArtifact(t, cfg)
	scenario, omitted := ScenarioFromArtifact(original)
	if omitted != 0 {
		t.Fatalf("a hand-driven run holds only scenario commands, %d omitted", omitted)
	}
	replayCfg := ScenarioConfig(scenario, spec, a, model.SinkInproc, "")
	if err := CheckScenario(context.Background(), replayCfg, scenario); err != nil {
		t.Fatalf("the exported scenario fails its check: %v", err)
	}
	reproduced := executedArtifact(t, replayCfg, scenario)
	if reproduced.ExpectedTraceDigest != original.ExpectedTraceDigest || reproduced.WorldDigest != original.WorldDigest {
		t.Fatalf("trace %s / world %s, want %s / %s", reproduced.ExpectedTraceDigest, reproduced.WorldDigest,
			original.ExpectedTraceDigest, original.WorldDigest)
	}
	if got, want := commandLogDigest(t, reproduced), commandLogDigest(t, original); got != want {
		t.Fatalf("command log digest %s, want %s", got, want)
	}
}

func executedArtifact(t *testing.T, cfg Config, scenario *model.Scenario) *model.RunArtifact {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ExecuteScenario(context.Background(), scenario); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

// commandLogDigest is the canonical digest of the command log as the artifact
// writes it.
func commandLogDigest(t *testing.T, art *model.RunArtifact) string {
	t.Helper()
	raw, err := json.Marshal(art.CommandLog)
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := model.DecodeBytes(raw, &document); err != nil {
		t.Fatal(err)
	}
	digest, err := canonical.Digest(document)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestCheckScenarioReportsEveryIssueAndOpensNoSink(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	trace := filepath.Join(t.TempDir(), "trace.jsonl")
	scenario := &model.Scenario{ID: "bad", World: model.ScenarioWorld{StartTimeNS: model.DefaultStartTimeNS, Profile: "rush"},
		Commands: []model.ScenarioCommand{{Op: model.OpPerturbApply, Args: map[string]any{"perturbation": "no-such", "from_ns": int64(0), "until_ns": int64(0)}}}}
	err := CheckScenario(context.Background(), ScenarioConfig(scenario, spec, a, model.SinkFile, trace), scenario)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractScenario || len(invalid.Issues) != 2 {
		t.Fatalf("CheckScenario = %v, want a scenario violation with both issues", err)
	}
	if _, statErr := os.Stat(trace); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("the check opened the sink: %v", statErr)
	}
}

func TestNewRefusesAProfileTheDomainDoesNotDeclare(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	_, err := New(context.Background(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc, ScenarioProfile: "rush"})
	if !errors.Is(err, ErrUnknownProfile) {
		t.Fatalf("New = %v, want ErrUnknownProfile", err)
	}
}

func TestARefusalOtherThanAnEffectorStopsTheScenario(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	cfg := Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc}
	scenario := &model.Scenario{Commands: []model.ScenarioCommand{{Op: model.OpFaultClear, Args: map[string]any{"fault_id": "never", "at_ns": int64(0)}}}}
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = r.ExecuteScenario(context.Background(), scenario)
	if err == nil || !strings.Contains(err.Error(), "run: scenario command 0 (fault.clear)") {
		t.Fatalf("ExecuteScenario = %v, want the refusal of command 0 named", err)
	}
}
