package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type ScenarioHooks struct {
	Before func(ctx context.Context, index int, command model.ScenarioCommand) (stop bool, err error)
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
		if err := r.executeStep(ctx, index, step); err != nil {
			return ScenarioOutcome{Executed: index}, err
		}
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

func (r *Run) executeStep(ctx context.Context, index int, step model.ScenarioCommand) error {
	if err := ctx.Err(); err != nil {
		return scenarioStepError(index, step, err)
	}
	command := model.Command{Op: step.Op, Args: step.Args}
	return scenarioStepError(index, step, executeCommand(ctx, r, &command, keepWorldRefusals))
}

func scenarioStepError(index int, step model.ScenarioCommand, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("run: scenario command %d (%s): %w", index, step.Op, err)
}
