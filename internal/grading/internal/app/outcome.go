package app

import (
	"context"
	"fmt"
	"path/filepath"

	catalog "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

type OutcomeOptions struct {
	RunDir, DomainsDir, AdaptersDir string
	Workers                         int
}

func GradeOutcome(ctx context.Context, options OutcomeOptions) (*model.OutcomeReport, error) {
	art, err := blindArtifact(filepath.Join(options.RunDir, model.FileRunArtifact))
	if err != nil {
		return nil, err
	}
	spec, adap, err := replayInputs(Options{DomainsDir: options.DomainsDir, AdaptersDir: options.AdaptersDir}, art)
	if err != nil {
		return nil, err
	}
	if err := requireReplayed(ctx, art, spec, adap); err != nil {
		return nil, err
	}
	return scoreOutcome(ctx, art, outcomeInputs{spec, adap}, options.Workers)
}

func blindArtifact(path string) (*model.RunArtifact, error) {
	art, err := run.LoadArtifact(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return art, domain.RequireBlind(art)
}

func requireReplayed(ctx context.Context, art *model.RunArtifact, spec *catalog.Compiled, adap *model.Adapter) error {
	replayed, err := run.ReplayArtifact(ctx, art, spec, adap, "")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return domain.RequireMatch(replayed)
}

type outcomeInputs struct {
	spec    *catalog.Compiled
	adapter *model.Adapter
}

func scoreOutcome(ctx context.Context, art *model.RunArtifact, inputs outcomeInputs, workers int) (*model.OutcomeReport, error) {
	runs, err := replayOutcomes(ctx, domain.OutcomeModes(inputs.spec.Spec), workers, inputs.replayer(art))
	if err != nil {
		return nil, err
	}
	scored, err := score.Outcome(inputs.spec.Spec.GroundTruth.OutcomeWeights, runs)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return &model.OutcomeReport{RunID: art.RunID, Domain: art.Domain.ID, OutcomeScore: *scored}, nil
}

func (inputs outcomeInputs) replayer(art *model.RunArtifact) outcomeReplay {
	return func(ctx context.Context, mode model.OutcomeMode) (model.OutcomeRun, error) {
		facts, err := run.ReplayOutcome(ctx, art, inputs.spec, inputs.adapter, mode)
		if err != nil {
			return model.OutcomeRun{}, fmt.Errorf("streamsim: %s: %w", mode, err)
		}
		return model.OutcomeRun{Mode: mode, Facts: facts}, nil
	}
}
