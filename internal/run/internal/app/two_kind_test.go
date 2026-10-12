package app

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twoKindConfig(t *testing.T) Config {
	t.Helper()
	return configFor(t, testsupport.TwoKindDomain())
}

func twoKindScenario(start int64) *model.Scenario {
	inject := func(entity, fault string, at int64) model.ScenarioCommand {
		return model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{"entity_id": entity, "fault": fault, "onset_ns": at}}
	}
	invoke := func(effector, entity, argument string, at int64) model.ScenarioCommand {
		return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": effector, "entity_id": entity,
			"command_id": effector + "-" + entity, "args": map[string]any{argument: entity}, "at_ns": at}}
	}
	return &model.Scenario{Commands: []model.ScenarioCommand{
		inject("tank-1", "tank_leak", start+600e9), inject("pump-2", "pump_wear", start+900e9),
		invoke("refill", "tank-1", "tank_id", start+1800e9), invoke("boost", "pump-3", "pump_id", start+1800e9),
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 2*3600e9}},
	}}
}

func runTwoKind(t *testing.T) (*Run, *model.RunArtifact) {
	t.Helper()
	cfg := twoKindConfig(t)
	return runScenario(t, cfg, twoKindScenario(cfg.StartTimeNS))
}

func TestATwoKindRunIsByteReproducibleAndReplaysToTheSameDigest(t *testing.T) {
	t.Parallel()
	first, art := runTwoKind(t)
	second, _ := runTwoKind(t)
	if len(first.trace) == 0 || string(first.trace) != string(second.trace) {
		t.Fatal("two runs of the same two-kind scenario diverged")
	}
	if calls := first.World.EffectorCalls(); len(calls) != 2 || !calls[0].EffectApplied || !calls[1].EffectApplied {
		t.Fatalf("effector calls = %+v, want both kinds' effectors applied", calls)
	}
	res, err := ReplayArtifact(context.Background(), art, nil, nil, "")
	if err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want a matching digest", res, err)
	}
}

func TestExplicitEntityIdsMayNameEveryInstanceSet(t *testing.T) {
	t.Parallel()
	cfg := twoKindConfig(t)
	cfg.EntityIDs = []string{"tank-1", "pump-1", "pump-2"}
	if ids := startRun(t, cfg).World.InitialEntityIDs(); len(ids) != 3 {
		t.Fatalf("entities = %v", ids)
	}
	cfg.EntityIDs = []string{"valve-1"}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "does not fit instance id template") {
		t.Fatalf("New = %v, want a refusal of an id that fits no instance set", err)
	}
}
