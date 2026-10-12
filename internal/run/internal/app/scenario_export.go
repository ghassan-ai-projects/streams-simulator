package app

import (
	"context"
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// PruneRefusedCalls is the scenario without the effector calls the run
// refuses at admission (world.ErrInvalidInvocation), and how many it dropped.
// A recorded run keeps a consumer's call to an entity that has gone, or with
// arguments the effector refuses; the world refused it before recording or
// drawing anything, so the scenario without it reproduces the same trace and
// world. Every other command is kept, for the scenario check to judge.
func PruneRefusedCalls(ctx context.Context, cfg Config, scenario *model.Scenario) (*model.Scenario, int, error) {
	shadow, err := newShadow(ctx, cfg)
	if err != nil {
		return nil, 0, err
	}
	pruned := *scenario
	pruned.Commands = []model.ScenarioCommand{}
	for _, step := range scenario.Commands {
		if !refusedAtAdmission(ctx, shadow, step) {
			pruned.Commands = append(pruned.Commands, step)
		}
	}
	return &pruned, len(scenario.Commands) - len(pruned.Commands), nil
}

// refusedAtAdmission executes one command on the shadow run and reports an
// effector call the world refused at admission.
func refusedAtAdmission(ctx context.Context, shadow *Run, step model.ScenarioCommand) bool {
	command := model.Command{Op: step.Op, Args: step.Args}
	return errors.Is(executeCommand(ctx, shadow, &command, keepWorldRefusals), world.ErrInvalidInvocation)
}

// CheckArtifactInputs reports whether the domain and adapter are the ones a
// run artifact recorded, by id, version and digest; a difference wraps
// ErrDigestMismatch.
func CheckArtifactInputs(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) error {
	return validateReplayInputs(art, spec, adapterSpec)
}
