package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func interlockScenario() *model.Scenario {
	start := model.DefaultStartTimeNS
	return &model.Scenario{ID: "after", World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": "line-1", "fault": "guard_trigger", "onset_ns": start, "params": map[string]any{}}},
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 60e9}},
		{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "reset_line", "entity_id": "line-1", "command_id": "c-1", "args": map[string]any{"line_id": "line-1"}, "at_ns": start + 60e9}},
		{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "adjust_speed", "entity_id": "line-1", "command_id": "c-2", "args": map[string]any{"line_id": "line-1"}, "at_ns": start + 60e9}},
		{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "adjust_speed", "entity_id": "line-1", "command_id": "c-2", "args": map[string]any{"line_id": "line-1", "speed": 3.0}, "at_ns": start + 60e9}},
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 120e9}},
	}}
}

func lineRun(t *testing.T, scenario *model.Scenario) *Run {
	t.Helper()
	spec, err := domain.Load(filepath.Join(testsupport.DomainsDir(), "discrete-line-oee.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, native := testBase(t)
	r, err := New(t.Context(), ScenarioConfig(scenario, spec, native, model.SinkInproc, ""))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

type seenStep struct {
	index int
	op    string
	ack   *model.EffectorAcknowledgement
}

func TestTheAfterHookSeesEveryExecutedCommandWithTheEffectorAcknowledgement(t *testing.T) {
	t.Parallel()
	scenario := interlockScenario()
	r := lineRun(t, scenario)
	var seen []seenStep
	hooks := ScenarioHooks{After: func(_ context.Context, index int, command model.ScenarioCommand, result StepResult) {
		seen = append(seen, seenStep{index, command.Op, result.Acknowledgement})
	}}
	outcome, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks)
	if err != nil || outcome.Executed != 6 || len(seen) != 6 {
		t.Fatalf("outcome = %+v, %v, seen %d, want all 6 commands", outcome, err, len(seen))
	}
	for i, step := range seen {
		if step.index != i || (step.op == model.OpEffectorInvoke) != (step.ack != nil) {
			t.Fatalf("step %d = %+v: only effector commands carry an acknowledgement", i, step)
		}
	}
}

func TestAnInterlockRefusalIsARecordedOutcomeAndTheScenarioGoesOn(t *testing.T) {
	t.Parallel()
	scenario := interlockScenario()
	r := lineRun(t, scenario)
	acks := map[int]*model.EffectorAcknowledgement{}
	hooks := ScenarioHooks{After: func(_ context.Context, index int, _ model.ScenarioCommand, result StepResult) {
		acks[index] = result.Acknowledgement
	}}
	if _, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks); err != nil {
		t.Fatalf("a refused effector call must not abort the scenario: %v", err)
	}
	refused := acks[2]
	if refused == nil || refused.Accepted || refused.Reason != "interlock_refused" || refused.CommandID != "c-1" {
		t.Fatalf("interlock refusal = %+v, want a recorded refusal", refused)
	}
	if reused := acks[4]; reused == nil || reused.Accepted || reused.Reason != "command_id_reused" {
		t.Fatalf("reused command id = %+v, want a recorded refusal", reused)
	}
	if first := acks[3]; first == nil || first.CommandID != "c-2" {
		t.Fatalf("first c-2 call = %+v, want an acknowledgement", first)
	}
	art, err := r.End("")
	if err != nil || len(art.CommandLog) != 6 {
		t.Fatalf("End = %d commands, %v; refused calls stay in the command log", len(art.CommandLog), err)
	}
}

func TestTheAfterHookSeesTheIdTheWorldGaveAnInjectedFault(t *testing.T) {
	t.Parallel()
	scenario := interlockScenario()
	start := model.DefaultStartTimeNS
	scenario.Commands = append(scenario.Commands,
		model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{"entity_id": "line-1", "fault": "guard_trigger", "onset_ns": start + 120e9, "params": map[string]any{}}},
		model.ScenarioCommand{Op: model.OpFaultClear, Args: map[string]any{"fault_id": "f-0", "at_ns": start + 120e9}})
	r := lineRun(t, scenario)
	ids := map[int]string{}
	hooks := ScenarioHooks{After: func(_ context.Context, index int, _ model.ScenarioCommand, result StepResult) {
		ids[index] = result.FaultID
	}}
	if _, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks); err != nil {
		t.Fatal(err)
	}
	if ids[0] != "f-0" || ids[6] != "f-1" || ids[7] != "" || ids[1] != "" {
		t.Fatalf("fault ids by step = %v, want f-0 at 0, f-1 at 6 and none elsewhere", ids)
	}
}
