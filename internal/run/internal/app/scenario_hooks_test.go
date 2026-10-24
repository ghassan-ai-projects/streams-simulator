package app

import (
	"context"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func advanceScenario(steps int) *model.Scenario {
	start := model.DefaultStartTimeNS
	commands := make([]model.ScenarioCommand, 0, steps)
	for i := 1; i <= steps; i++ {
		commands = append(commands, model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + int64(i)*60e9}})
	}
	return &model.Scenario{ID: "hooks", World: model.ScenarioWorld{StartTimeNS: start}, Commands: commands}
}

func hookConfig(t *testing.T, scenario *model.Scenario) Config {
	t.Helper()
	spec, a := testBase(t)
	return ScenarioConfig(scenario, spec, a, model.SinkInproc, "")
}

func hookRun(t *testing.T, scenario *model.Scenario) *Run {
	t.Helper()
	r, err := New(t.Context(), hookConfig(t, scenario))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAHookThatStopsEndsTheScenarioAfterTheCommandsAlreadyExecuted(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(6)
	r := hookRun(t, scenario)
	seen := []int{}
	hooks := ScenarioHooks{Before: func(_ context.Context, index int, _ model.ScenarioCommand) (bool, error) {
		seen = append(seen, index)
		return index == 3, nil
	}}
	outcome, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks)
	if err != nil || outcome.Executed != 3 || !outcome.Stopped {
		t.Fatalf("outcome = %+v, %v, want 3 executed and stopped", outcome, err)
	}
	if len(seen) != 4 {
		t.Fatalf("the hook saw %v, want indexes 0..3", seen)
	}
	art, err := r.End("")
	if err != nil || len(art.CommandLog) != 3 {
		t.Fatalf("End = %d commands, %v, want exactly the 3 executed", len(art.CommandLog), err)
	}
	for i, command := range art.CommandLog {
		if got, want := command.Args["to_ns"], scenario.Commands[i].Args["to_ns"]; got != want {
			t.Fatalf("command %d to_ns = %v, want %v", i, got, want)
		}
	}
}

func TestAHookThatStopsBeforeTheFirstCommandRunsNothing(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(2)
	r := hookRun(t, scenario)
	outcome, err := r.ExecuteScenarioWith(t.Context(), scenario, ScenarioHooks{Before: func(context.Context, int, model.ScenarioCommand) (bool, error) {
		return true, nil
	}})
	if err != nil || outcome.Executed != 0 || !outcome.Stopped {
		t.Fatalf("outcome = %+v, %v", outcome, err)
	}
}

func TestAHookErrorStopsTheScenarioAndIsWrapped(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(4)
	r := hookRun(t, scenario)
	boom := errors.New("boom")
	outcome, err := r.ExecuteScenarioWith(t.Context(), scenario, ScenarioHooks{Before: func(_ context.Context, index int, _ model.ScenarioCommand) (bool, error) {
		if index == 2 {
			return true, boom
		}
		return false, nil
	}})
	if !errors.Is(err, boom) || outcome.Executed != 2 || outcome.Stopped {
		t.Fatalf("outcome = %+v, %v, want the hook error after 2 commands, not a clean stop", outcome, err)
	}
}

func TestACancelledContextStopsBeforeTheHookAndTheCommand(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(3)
	r := hookRun(t, scenario)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	_, err := r.ExecuteScenarioWith(ctx, scenario, ScenarioHooks{Before: func(context.Context, int, model.ScenarioCommand) (bool, error) {
		called = true
		return false, nil
	}})
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("err = %v, hook called = %v, want a cancellation before the hook", err, called)
	}
}

func TestTheZeroHookPathIsTheSameRunAsExecuteScenario(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(5)
	plain, hooked := hookRun(t, scenario), hookRun(t, scenario)
	if err := plain.ExecuteScenario(t.Context(), scenario); err != nil {
		t.Fatal(err)
	}
	outcome, err := hooked.ExecuteScenarioWith(t.Context(), scenario, ScenarioHooks{})
	if err != nil || outcome.Executed != 5 || outcome.Stopped {
		t.Fatalf("outcome = %+v, %v", outcome, err)
	}
	a, errA := plain.End("")
	b, errB := hooked.End("")
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}
	if a.ExpectedTraceDigest != b.ExpectedTraceDigest || a.WorldDigest != b.WorldDigest || commandLogDigest(t, a) != commandLogDigest(t, b) {
		t.Fatalf("trace %s/%s world %s/%s: the zero-hook path changed the run", a.ExpectedTraceDigest, b.ExpectedTraceDigest, a.WorldDigest, b.WorldDigest)
	}
}

func TestTheHookRunsOutsideTheCommandLock(t *testing.T) {
	t.Parallel()
	scenario := advanceScenario(3)
	r := hookRun(t, scenario)
	var clocks []int64
	hooks := ScenarioHooks{Before: func(context.Context, int, model.ScenarioCommand) (bool, error) {
		clocks = append(clocks, r.Status().ClockNS)
		return false, nil
	}}
	if _, err := r.ExecuteScenarioWith(t.Context(), scenario, hooks); err != nil {
		t.Fatal(err)
	}
	if len(clocks) != 3 || clocks[0] >= clocks[2] {
		t.Fatalf("status clocks seen by the hook = %v, want the clock to advance between commands", clocks)
	}
}
