package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func pondInvocation(commandID string, level float64, atNS int64) model.ScenarioCommand {
	return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "start_aerator",
		"entity_id": "site-a/pond-1", "command_id": commandID, "args": map[string]any{"pond_id": "site-a/pond-1", "level": level}, "at_ns": atNS}}
}

func pondConfig(t *testing.T) Config {
	t.Helper()
	spec, a := testBase(t)
	return Config{Domain: spec, Adapter: a, Seed: 2, SinkName: model.SinkInproc, StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true}
}

// A command id reused for a different request is refused by the world as it
// runs: the refusal is recorded, the effect is not applied and the scenario
// goes on.
func TestAWorldRefusalIsRecordedAndTheScenarioGoesOn(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), pondConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	scenario := &model.Scenario{Commands: []model.ScenarioCommand{pondInvocation("same-id", 1, start), pondInvocation("same-id", 0.5, start),
		{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 600e9}}}}
	if err := r.ExecuteScenario(context.Background(), scenario); err != nil {
		t.Fatal(err)
	}
	if calls := r.World.EffectorCalls(); len(calls) != 1 || calls[0].CommandID != "same-id" {
		t.Fatalf("effector calls = %+v, want only the first invocation applied", calls)
	}
	if log := r.commandLog; len(log) != 3 || log[1].Op != model.OpEffectorInvoke || log[2].Op != model.OpClockAdvance {
		t.Fatalf("command log = %+v, want both invocations and the advance", log)
	}
}

// An invocation the world refuses because the command is wrong (here, an
// entity retired earlier) fails the scenario at that command.
func TestAWrongInvocationStopsTheScenario(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), pondConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	scenario := &model.Scenario{Commands: []model.ScenarioCommand{
		{Op: model.OpEntityRetire, Args: map[string]any{"entity_id": "site-a/pond-1", "reason": "drained", "at_ns": start}},
		pondInvocation("c-1", 1, start)}}
	err = r.ExecuteScenario(context.Background(), scenario)
	if err == nil || !strings.Contains(err.Error(), "run: scenario command 1 (effector.invoke)") || !strings.Contains(err.Error(), "unknown entity") {
		t.Fatalf("ExecuteScenario = %v, want command 1 refused for its entity", err)
	}
}

func TestOnlyReplayKeepsEveryEffectorRefusal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		err              error
		replay, authored bool
	}{
		{fmt.Errorf("InvokeEffector: %w", world.ErrInterlockRefused), true, true},
		{fmt.Errorf("InvokeEffector: %w: %q", world.ErrCommandIDReused, "c-1"), true, true},
		{errors.New(`world: unknown entity "x"`), true, false},
		{errors.New("run: InvokeEffector on a finished run"), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			t.Parallel()
			if keepEveryRefusal.keeps(tc.err) != tc.replay || keepWorldRefusals.keeps(tc.err) != tc.authored {
				t.Fatalf("replay keeps %v, authored keeps %v; want %v, %v", keepEveryRefusal.keeps(tc.err), keepWorldRefusals.keeps(tc.err), tc.replay, tc.authored)
			}
		})
	}
}

// A director-driven run (an environment injection, an advance that awaited a
// consumer) exports to a scenario that reproduces its trace and world; the
// command log differs only by the omitted entry and the awaited flag.
func TestADirectorRunExportsToAScenarioThatReproducesIt(t *testing.T) {
	t.Parallel()
	cfg := pondConfig(t)
	start := cfg.StartTimeNS
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	r.ConfigureEnvTarget("consumer-1", true)
	r.SetQuiesceParkedHook(func() { r.ReportQuiesced(start + 3600e9) })
	steps := []error{
		firstError(r.EnvInject("consumer-1", "pause", nil, start)),
		firstError(r.Advance(context.Background(), start+3600e9, true)),
		firstError(r.InvokeEffector("start_aerator", "site-a/pond-1", "c-1", map[string]any{"pond_id": "site-a/pond-1", "level": 1.0}, start+3600e9)),
		firstError(r.Advance(context.Background(), start+7200e9, false)),
	}
	if err := errors.Join(steps...); err != nil {
		t.Fatal(err)
	}
	original, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	scenario, omitted := ScenarioFromArtifact(original)
	if omitted != 1 || len(scenario.Commands) != 3 {
		t.Fatalf("scenario = %+v, omitted %d; want three commands and the env.inject omitted", scenario.Commands, omitted)
	}
	reproduced := executedArtifact(t, ScenarioConfig(scenario, cfg.Domain, cfg.Adapter, model.SinkInproc, ""), scenario)
	if reproduced.ExpectedTraceDigest != original.ExpectedTraceDigest || reproduced.WorldDigest != original.WorldDigest {
		t.Fatalf("trace %s / world %s, want %s / %s", reproduced.ExpectedTraceDigest, reproduced.WorldDigest, original.ExpectedTraceDigest, original.WorldDigest)
	}
	if awaited := reproduced.CommandLog[0].Args["await_consumer"]; awaited != false {
		t.Fatalf("reproduced advance await_consumer = %v, want false", awaited)
	}
}

func TestScenarioConfigCarriesEveryWorldSetting(t *testing.T) {
	t.Parallel()
	cfg := pondConfig(t)
	scenario := &model.Scenario{Seed: 9, World: model.ScenarioWorld{StartTimeNS: 42, Profile: "p", EntityIDs: []string{"a", "b"},
		Noiseless: true, ForceFailureMode: "reject", ClockMultiplier: 3}}
	got := ScenarioConfig(scenario, cfg.Domain, cfg.Adapter, model.SinkFile, "t.jsonl")
	if got.Seed != 9 || got.StartTimeNS != 42 || !got.StartTimeSet || got.ScenarioProfile != "p" || len(got.EntityIDs) != 2 ||
		!got.Noiseless || got.ForceFailureMode != world.ModeReject || got.ClockMultiplier != 3 ||
		got.SinkName != model.SinkFile || got.SinkTarget != "t.jsonl" || got.TimeMode != model.TimeStepped {
		t.Fatalf("config = %+v, want every scenario setting", got)
	}
}

// A world the scenario's settings cannot create is a violation located at
// the world settings.
func TestAWorldTheSettingsCannotCreateIsAViolation(t *testing.T) {
	t.Parallel()
	cfg := pondConfig(t)
	scenario := &model.Scenario{ID: "bad", World: model.ScenarioWorld{StartTimeNS: cfg.StartTimeNS, ForceFailureMode: "bogus_mode"}}
	err := CheckScenario(context.Background(), ScenarioConfig(scenario, cfg.Domain, cfg.Adapter, model.SinkInproc, ""), scenario)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/world" ||
		!strings.Contains(invalid.Issues[0].Message, `unknown failure mode "bogus_mode"`) {
		t.Fatalf("CheckScenario = %+v, want one issue at /world naming the mode", err)
	}
}

func TestAScenarioWithEntityIdsOutsideTheDomainInstancesIsAViolation(t *testing.T) {
	t.Parallel()
	cfg := pondConfig(t)
	scenario := &model.Scenario{ID: "bad-ids", World: model.ScenarioWorld{StartTimeNS: cfg.StartTimeNS, EntityIDs: []string{"site-a/pond-1", "tank-9"}}}
	err := CheckScenario(context.Background(), ScenarioConfig(scenario, cfg.Domain, cfg.Adapter, model.SinkInproc, ""), scenario)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/world" ||
		!strings.Contains(invalid.Issues[0].Message, `entity id "tank-9" does not fit instance id template`) {
		t.Fatalf("CheckScenario = %+v, want one issue at /world naming the id", err)
	}
	cfg.EntityIDs = []string{"tank-9"}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "does not fit instance id template") {
		t.Fatalf("New = %v, want a refusal of the entity id outside the instances", err)
	}
}
