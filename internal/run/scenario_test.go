package run_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// The facade checks a scenario, runs it, and exports the run back to the
// same scenario.
func TestAScenarioRunsAndExportsThroughTheFacade(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	scenario := &model.Scenario{ID: "facade", Seed: 3, World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": "site-a/pond-1", "fault": "aerator_failure", "onset_ns": start + 60e9, "params": nil}},
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 3600e9}}}}
	cfg := run.ScenarioConfig(scenario, spec, a, model.SinkInproc, "")
	if err := run.CheckScenario(context.Background(), cfg, scenario); err != nil {
		t.Fatal(err)
	}
	r, err := run.New(context.Background(), cfg)
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
	exported, omitted := run.ScenarioFromArtifact(art)
	if omitted != 0 || len(exported.Commands) != 2 || exported.Seed != 3 || exported.World.StartTimeNS != start {
		t.Fatalf("exported = %+v, omitted %d; want the scenario back", exported, omitted)
	}
	broken := &model.Scenario{ID: "broken", World: model.ScenarioWorld{StartTimeNS: start, Profile: "no-such-profile"}}
	var invalid *model.ValidationError
	if err := run.CheckScenario(context.Background(), run.ScenarioConfig(broken, spec, a, model.SinkInproc, ""), broken); !errors.As(err, &invalid) || invalid.Issues[0].Pointer != "/world/profile" {
		t.Fatalf("CheckScenario = %v, want the profile located", err)
	}
}
