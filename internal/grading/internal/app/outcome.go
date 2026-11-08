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

type OutcomeOptions struct{ RunDir, DomainsDir, AdaptersDir string }

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
	return scoreOutcome(ctx, art, spec, adap)
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

func scoreOutcome(ctx context.Context, art *model.RunArtifact, spec *catalog.Compiled, adap *model.Adapter) (*model.OutcomeReport, error) {
	var runs []model.OutcomeRun
	for _, mode := range domain.OutcomeModes(spec.Spec) {
		facts, err := run.ReplayOutcome(ctx, art, spec, adap, mode)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %s: %w", mode, err)
		}
		runs = append(runs, model.OutcomeRun{Mode: mode, Facts: facts})
	}
	scored, err := score.Outcome(spec.Spec.GroundTruth.OutcomeWeights, runs)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return &model.OutcomeReport{RunID: art.RunID, Domain: art.Domain.ID, OutcomeScore: *scored}, nil
}
