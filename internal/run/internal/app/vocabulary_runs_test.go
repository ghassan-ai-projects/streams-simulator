package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type fixtureRun struct {
	name, domain, entity, fault, effector, argument string
}

func (f fixtureRun) run(t *testing.T) (*Run, *model.RunArtifact) {
	t.Helper()
	cfg := configFor(t, f.domain)
	start := cfg.StartTimeNS
	return runScenario(t, cfg, &model.Scenario{Commands: []model.ScenarioCommand{
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": f.entity, "fault": f.fault, "onset_ns": start + 1800e9}},
		{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": f.effector, "entity_id": f.entity, "command_id": "c-1",
			"args": map[string]any{f.argument: f.entity}, "at_ns": start + 3600e9}},
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 2*3600e9}},
	}})
}

func TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest(t *testing.T) {
	t.Parallel()
	for _, f := range []fixtureRun{
		{"transfer", testsupport.TransferDomain(), "cell-1", "pipe_blockage", "speed_up", "cell_id"},
		{"utility", testsupport.UtilityDomain(), "meter-1", "load_surge", "shed_load", "meter_id"},
		{"chiller", testsupport.UtilityDomain(), "CH1", "flow_loss", "start_chiller", "chiller_id"},
	} {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			first, art := f.run(t)
			second, _ := f.run(t)
			if len(first.trace) == 0 || string(first.trace) != string(second.trace) {
				t.Fatal("two runs of the same scenario diverged")
			}
			res, err := ReplayArtifact(context.Background(), art, nil, nil, "")
			if err != nil || !res.Matches {
				t.Fatalf("replay = %+v, %v, want a matching digest", res, err)
			}
		})
	}
}
