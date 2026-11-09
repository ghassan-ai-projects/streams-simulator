package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

// withDoNothing adds the run replayed with every effector call removed, the truth a failure prediction is graded
// against, when the verdict carries one.
func withDoNothing(ctx context.Context, options Options, art *model.RunArtifact, evidence score.Evidence) (score.Evidence, error) {
	if !evidence.Verdict.PredictsFailure() {
		return evidence, nil
	}
	spec, adap, err := replayInputs(options, art)
	if err != nil {
		return score.Evidence{}, err
	}
	evidence.DoNothing, err = run.ReplayOutcome(ctx, art, spec, adap, model.OutcomeDoNothing)
	if err != nil {
		return score.Evidence{}, fmt.Errorf("streamsim: %s: %w", model.OutcomeDoNothing, err)
	}
	return evidence, nil
}
