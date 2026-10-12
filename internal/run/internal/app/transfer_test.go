package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func runTransfer(t *testing.T) (*Run, *model.RunArtifact) {
	t.Helper()
	cfg := configFor(t, testsupport.TransferDomain())
	start := cfg.StartTimeNS
	return runScenario(t, cfg, &model.Scenario{Commands: []model.ScenarioCommand{
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": "cell-1", "fault": "pipe_blockage", "onset_ns": start + 1800e9}},
		{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "speed_up", "entity_id": "cell-1", "command_id": "c-1",
			"args": map[string]any{"cell_id": "cell-1"}, "at_ns": start + 3600e9}},
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 2*3600e9}},
	}})
}

func TestATransferRunIsByteReproducibleAndReplaysToTheSameDigest(t *testing.T) {
	t.Parallel()
	first, art := runTransfer(t)
	second, _ := runTransfer(t)
	if len(first.trace) == 0 || string(first.trace) != string(second.trace) {
		t.Fatal("two runs of the same transfer scenario diverged")
	}
	res, err := ReplayArtifact(context.Background(), art, nil, nil, "")
	if err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want a matching digest", res, err)
	}
}
