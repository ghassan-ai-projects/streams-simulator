package app

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func structuredConfig(t *testing.T) Config {
	t.Helper()
	return configFor(t, testsupport.StructuredDomain())
}

func configFor(t *testing.T, path string) Config {
	t.Helper()
	spec, err := domain.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return Config{Domain: spec, Adapter: native, Seed: 9, SinkName: model.SinkInproc, StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true}
}

func structuredScenario(start int64) *model.Scenario {
	inject := func(entity, fault string, at int64) model.ScenarioCommand {
		return model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{"entity_id": entity, "fault": fault, "onset_ns": at}}
	}
	invoke := func(effector, entity string, args map[string]any, at int64) model.ScenarioCommand {
		return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": effector, "entity_id": entity,
			"command_id": effector + "-" + entity, "args": args, "at_ns": at}}
	}
	return &model.Scenario{Commands: []model.ScenarioCommand{
		inject("T-main", "tank_overheat", start+600e9), inject("tank-1", "tank_leak", start+900e9),
		invoke("chill", "T-main", map[string]any{"tank_id": "T-main"}, start+1800e9),
		invoke("boost", "P-slow", map[string]any{"pump_id": "P-slow", "speed": 40.0}, start+1800e9),
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 2*3600e9}},
	}}
}

func runStructured(t *testing.T) (*Run, *model.RunArtifact) {
	t.Helper()
	cfg := structuredConfig(t)
	return runScenario(t, cfg, structuredScenario(cfg.StartTimeNS))
}

func runScenario(t *testing.T, cfg Config, scenario *model.Scenario) (*Run, *model.RunArtifact) {
	t.Helper()
	r := startRun(t, cfg)
	if err := r.ExecuteScenario(context.Background(), scenario); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return r, art
}

func startRun(t *testing.T, cfg Config) *Run {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestARunWithOverridesIsByteReproducibleAndReplaysToTheSameDigest(t *testing.T) {
	t.Parallel()
	first, art := runStructured(t)
	second, _ := runStructured(t)
	if len(first.trace) == 0 || string(first.trace) != string(second.trace) {
		t.Fatal("two runs of the same scenario diverged")
	}
	if !strings.Contains(string(first.trace), `"tank.temp"`) {
		t.Fatal("the added channel never reached the trace")
	}
	if calls := first.World.EffectorCalls(); len(calls) != 2 || !calls[0].EffectApplied || !calls[1].EffectApplied {
		t.Fatalf("effector calls = %+v, want both added and narrowed effectors applied", calls)
	}
	res, err := ReplayArtifact(context.Background(), art, nil, nil, "")
	if err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want a matching digest", res, err)
	}
}

func TestStateHistoryOfAnOverriddenEntityIncludesItsAddedStates(t *testing.T) {
	t.Parallel()
	r, _ := runStructured(t)
	var named, plain bool
	for _, snapshot := range r.history {
		_, hasTemp := snapshot.States["temp"]
		named = named || (snapshot.Entity == "T-main" && hasTemp)
		plain = plain || (snapshot.Entity == "tank-1" && hasTemp)
	}
	if !named || plain {
		t.Fatalf("temp in history of T-main %v and of tank-1 %v, want true and false", named, plain)
	}
}

func TestExplicitEntityIdsMayNameNamedInstances(t *testing.T) {
	t.Parallel()
	cfg := structuredConfig(t)
	cfg.EntityIDs = []string{"T-main", "pump-1", "air_header"}
	r := startRun(t, cfg)
	if ids := r.World.InitialEntityIDs(); len(ids) != 3 || r.World.Entity("T-main").Type != "tank" {
		t.Fatalf("entities = %v", ids)
	}
	cfg.EntityIDs = []string{"T-other"}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "does not fit instance id template") {
		t.Fatalf("New = %v, want a refusal of an unknown name", err)
	}
}
