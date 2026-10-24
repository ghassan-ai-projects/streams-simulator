package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type ScenarioHooks struct {
	Before func(ctx context.Context, index int, command model.ScenarioCommand) (stop bool, err error)
	After  func(ctx context.Context, index int, command model.ScenarioCommand, result StepResult)
}

type ScenarioOutcome struct {
	Executed int
	Stopped  bool
}

func (r *Run) ExecuteScenario(ctx context.Context, scenario *model.Scenario) error {
	_, err := r.ExecuteScenarioWith(ctx, scenario, ScenarioHooks{})
	return err
}

func (r *Run) ExecuteScenarioWith(ctx context.Context, scenario *model.Scenario, hooks ScenarioHooks) (ScenarioOutcome, error) {
	for index, step := range scenario.Commands {
		if stop, err := hooks.before(ctx, index, step); err != nil || stop {
			return ScenarioOutcome{Executed: index, Stopped: stop}, err
		}
		result, err := r.executeStep(ctx, index, step)
		if err != nil {
			return ScenarioOutcome{Executed: index}, err
		}
		hooks.after(ctx, index, step, result)
	}
	return ScenarioOutcome{Executed: len(scenario.Commands)}, nil
}

func (hooks ScenarioHooks) before(ctx context.Context, index int, step model.ScenarioCommand) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, scenarioStepError(index, step, err)
	}
	if hooks.Before == nil {
		return false, nil
	}
	stop, err := hooks.Before(ctx, index, step)
	if err != nil {
		return false, fmt.Errorf("run: scenario hook before command %d (%s): %w", index, step.Op, err)
	}
	return stop, nil
}

func (hooks ScenarioHooks) after(ctx context.Context, index int, step model.ScenarioCommand, result StepResult) {
	if hooks.After != nil {
		hooks.After(ctx, index, step, result)
	}
}

func (r *Run) executeStep(ctx context.Context, index int, step model.ScenarioCommand) (StepResult, error) {
	if err := ctx.Err(); err != nil {
		return StepResult{}, scenarioStepError(index, step, err)
	}
	command := model.Command{Op: step.Op, Args: step.Args}
	result, err := executeObserved(ctx, r, &command, keepWorldRefusals)
	return result, scenarioStepError(index, step, err)
}

func scenarioStepError(index int, step model.ScenarioCommand, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("run: scenario command %d (%s): %w", index, step.Op, err)
}
