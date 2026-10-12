package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// churnConfig is the k8s cluster, whose pods are born and die as the run
// unfolds: by an hour in, seed 1 has born cluster-a/pod-100.
func churnConfig(t *testing.T) Config {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("k8s-cluster"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return Config{Domain: spec, Adapter: a, Seed: 1, SinkName: model.SinkInproc, StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true}
}

const churnBorn = "cluster-a/pod-100"

func advanceTo(ns int64) model.ScenarioCommand {
	return model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": ns}}
}

// refusedAt asserts a dry-run refusal located at one command.
func refusedAt(t *testing.T, err error, pointer, message string) {
	t.Helper()
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != pointer ||
		!strings.Contains(invalid.Issues[0].Message, message) {
		t.Fatalf("CheckScenario = %+v, want one issue at %s containing %q", err, pointer, message)
	}
}

// An entity the world births as it runs may be named once it exists; the
// check follows the run instead of the initial world.
func TestAnEntityBornDuringTheRunMayBeNamed(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	start := cfg.StartTimeNS
	scenario := &model.Scenario{ID: "churn", World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
		advanceTo(start + 3600e9),
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": churnBorn, "fault": "crash_loop", "onset_ns": start + 3600e9, "params": nil}},
		advanceTo(start + 7200e9)}}
	if err := CheckScenario(context.Background(), cfg, scenario); err != nil {
		t.Fatalf("CheckScenario = %v, want the churn-born entity accepted", err)
	}
	if art := executedArtifact(t, cfg, scenario); len(art.CommandLog) != 3 {
		t.Fatalf("command log = %+v", art.CommandLog)
	}
}

// What the run would refuse as it unfolds is found before anything is
// delivered, located at the command.
func TestTheDryRunFindsWhatTheRunWouldRefuse(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	start := cfg.StartTimeNS
	cases := []struct {
		name     string
		commands []model.ScenarioCommand
		pointer  string
		message  string
	}{
		{"adding an id the churn already born", []model.ScenarioCommand{advanceTo(start + 3600e9),
			{Op: model.OpEntityAdd, Args: map[string]any{"entity_id": churnBorn, "at_ns": start + 3600e9}}}, "/commands/1", "already exists"},
		{"an entity that is never born", []model.ScenarioCommand{
			{Op: model.OpFaultInject, Args: map[string]any{"entity_id": "cluster-a/pod-9999", "fault": "crash_loop", "onset_ns": start, "params": nil}}},
			"/commands/0", "unknown entity"},
		{"clearing a fault never injected", []model.ScenarioCommand{
			{Op: model.OpFaultClear, Args: map[string]any{"fault_id": "f-7", "at_ns": start}}}, "/commands/0", `unknown fault id "f-7"`},
		{"clearing a perturbation never applied", []model.ScenarioCommand{
			{Op: model.OpPerturbClear, Args: map[string]any{"perturb_id": "drop-3"}}}, "/commands/0", `unknown perturbation id "drop-3"`},
		{"an entity born before the clock", []model.ScenarioCommand{
			{Op: model.OpEntityAdd, Args: map[string]any{"entity_id": "late", "at_ns": int64(0)}}}, "/commands/0", "before the world clock"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			scenario := &model.Scenario{ID: "refused", World: model.ScenarioWorld{StartTimeNS: start}, Commands: tc.commands}
			refusedAt(t, CheckScenario(context.Background(), cfg, scenario), tc.pointer, tc.message)
		})
	}
}

// Every surface refuses an advance beyond the bound on a run's span.
func TestARunRefusesToAdvanceBeyondItsSpan(t *testing.T) {
	t.Parallel()
	r, err := New(context.Background(), pondConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), r.Config.StartTimeNS+MaxRunSpanNS+1, false); !errors.Is(err, ErrRunSpanExceeded) {
		t.Fatalf("Advance beyond the span = %v, want ErrRunSpanExceeded", err)
	}
	if r.World.Clock() != r.Config.StartTimeNS {
		t.Fatalf("the refused advance moved the clock to %d", r.World.Clock())
	}
}

// A director-driven run on a churning world exports to a scenario that passes
// its check and reproduces the run.
func TestAChurningRunExportsToAScenarioThatReproducesIt(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	start := cfg.StartTimeNS
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	steps := []error{
		firstError(r.Advance(context.Background(), start+3600e9, false)),
		firstError(r.InjectFault(churnBorn, "crash_loop", start+3600e9, nil)),
		firstError(r.ApplyPerturb("drop", nil, start+3700e9, start+3800e9)),
		firstError(r.Advance(context.Background(), start+7200e9, false)),
	}
	if err := errors.Join(steps...); err != nil {
		t.Fatal(err)
	}
	original, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	scenario, _ := ScenarioFromArtifact(original)
	scenarioCfg := ScenarioConfig(scenario, cfg.Domain, cfg.Adapter, model.SinkInproc, "")
	if err := CheckScenario(context.Background(), scenarioCfg, scenario); err != nil {
		t.Fatalf("the export of a churning run fails its check: %v", err)
	}
	if reproduced := executedArtifact(t, scenarioCfg, scenario); reproduced.ExpectedTraceDigest != original.ExpectedTraceDigest {
		t.Fatalf("trace %s, want %s", reproduced.ExpectedTraceDigest, original.ExpectedTraceDigest)
	}
}
