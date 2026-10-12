// Package app is the use-case layer of the grading module: it reads the files a
// run published, replays the artifact and scores the evidence.
package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

type Options struct{ Artifact, Label, DomainsDir, AdaptersDir string }

func Grade(ctx context.Context, options Options) (*score.Scorecard, error) {
	art, err := run.LoadArtifact(options.Artifact)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := domain.RequireBlind(art); err != nil {
		return nil, err
	}
	return gradeArtifact(ctx, options, art)
}

func gradeArtifact(ctx context.Context, options Options, art *model.RunArtifact) (*score.Scorecard, error) {
	dir := filepath.Dir(options.Artifact)
	verdict, gt, err := loadReports(dir, options.Label)
	if err != nil {
		return nil, err
	}
	evidence, err := scoringEvidence(ctx, options, art, dir)
	if err != nil {
		return nil, err
	}
	evidence.Verdict = verdict
	return gradeRun(evidence, gt)
}

func gradeRun(evidence score.Evidence, gt *model.GroundTruthRecord) (*score.Scorecard, error) {
	card, err := score.Score(evidence, gt)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return card, nil
}
