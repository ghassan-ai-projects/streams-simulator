package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestExportKeepsScenarioCommandsInOrderAndCountsTheRest(t *testing.T) {
	t.Parallel()
	art := &model.RunArtifact{RunID: "r-abc", Seed: 4, Domain: model.ArtifactRef{ID: "d"}, Adapter: model.ArtifactRef{ID: "a"},
		WorldConfig: model.WorldConfig{StartTimeNS: 7, EntityIDs: []string{"e-1"}, ScenarioProfile: "nominal", Noiseless: true,
			ForceFailureMode: "reject", ClockMultiplier: 2},
		CommandLog: []model.Command{
			{Op: model.OpWorldCreate}, {Op: model.OpFaultInject, Args: map[string]any{"fault": "f-1"}},
			{Op: model.OpEnvInject}, {Op: model.OpClockAdvance, Args: map[string]any{"to_ns": 9, "await_consumer": true}}, {Op: model.OpRunEnd},
		}}
	scenario, omitted := ScenarioFromArtifact(art)
	if omitted != 3 || len(scenario.Commands) != 2 || scenario.Commands[0].Op != model.OpFaultInject || scenario.Commands[1].Op != model.OpClockAdvance {
		t.Fatalf("scenario commands = %+v, omitted %d; want fault then advance, 3 omitted", scenario.Commands, omitted)
	}
	want := model.ScenarioWorld{StartTimeNS: 7, EntityIDs: []string{"e-1"}, Profile: "nominal", Noiseless: true, ForceFailureMode: "reject", ClockMultiplier: 2}
	got := scenario.World
	if scenario.ID != "r-abc" || scenario.Seed != 4 || got.StartTimeNS != want.StartTimeNS || got.Profile != want.Profile ||
		got.Noiseless != want.Noiseless || got.ForceFailureMode != want.ForceFailureMode || got.ClockMultiplier != want.ClockMultiplier ||
		len(got.EntityIDs) != 1 || got.EntityIDs[0] != "e-1" {
		t.Fatalf("scenario = %+v, want the artifact's inputs %+v", scenario, want)
	}
	if _, kept := scenario.Commands[1].Args["await_consumer"]; kept || len(scenario.Commands[1].Args) != 1 {
		t.Fatalf("advance args = %v, want only to_ns", scenario.Commands[1].Args)
	}
}

// The exported scenario owns its arguments: changing it leaves the artifact
// as recorded.
func TestAnExportedScenarioSharesNothingWithTheArtifact(t *testing.T) {
	t.Parallel()
	params := map[string]any{"severity": 1.0, "list": []any{map[string]any{"k": "v"}}}
	art := &model.RunArtifact{WorldConfig: model.WorldConfig{EntityIDs: []string{"e-1"}},
		CommandLog: []model.Command{{Op: model.OpFaultInject, Args: map[string]any{"fault": "f-1", "params": params}}}}
	scenario, _ := ScenarioFromArtifact(art)
	exported := AsMap(scenario.Commands[0].Args["params"])
	exported["severity"] = 2.0
	AsMap(exported["list"].([]any)[0])["k"] = "changed"
	scenario.Commands[0].Args["fault"] = "f-2"
	scenario.World.EntityIDs[0] = "e-2"
	recorded := art.CommandLog[0].Args
	if recorded["fault"] != "f-1" || params["severity"] != 1.0 || AsMap(params["list"].([]any)[0])["k"] != "v" || art.WorldConfig.EntityIDs[0] != "e-1" {
		t.Fatalf("changing the export changed the artifact: %+v, %v", recorded, art.WorldConfig.EntityIDs)
	}
}

func TestAnExportedScenarioIDFallsBackWhenTheRunIDIsNotOne(t *testing.T) {
	t.Parallel()
	for runID, want := range map[string]string{"r-1": "r-1", "Run_1": "exported-run", "": "exported-run"} {
		if got := exportedScenarioID(runID); got != want {
			t.Errorf("exportedScenarioID(%q) = %q, want %q", runID, got, want)
		}
	}
}

func TestOnlyDirectorOperationsAreScenarioCommands(t *testing.T) {
	t.Parallel()
	for op, want := range map[string]bool{model.OpClockAdvance: true, model.OpFaultClear: true, model.OpEntityRetire: true,
		model.OpWorldCreate: false, model.OpEnvInject: false, model.OpRunBegin: false, model.OpClockRun: false} {
		if got := IsScenarioOp(op); got != want {
			t.Errorf("IsScenarioOp(%s) = %v, want %v", op, got, want)
		}
	}
}
