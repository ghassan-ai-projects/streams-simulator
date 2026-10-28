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

func GradeRun(ctx context.Context, options Options) (*score.RunScore, error) {
	art, err := run.LoadArtifact(options.Artifact)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := domain.RequireBlind(art); err != nil {
		return nil, err
	}
	return gradeArtifact(ctx, options, art)
}

func Grade(ctx context.Context, options Options) (*score.Scorecard, error) {
	graded, err := GradeRun(ctx, options)
	if err != nil {
		return nil, err
	}
	if len(graded.Scorecards) != 1 {
		return nil, fmt.Errorf("streamsim: %w: the run holds %d", domain.ErrNotOneLabel, len(graded.Scorecards))
	}
	return graded.Scorecards[0], nil
}

func gradeArtifact(ctx context.Context, options Options, art *model.RunArtifact) (*score.RunScore, error) {
	dir := filepath.Dir(options.Artifact)
	verdict, labels, err := loadReports(dir, options.Label)
	if err != nil {
		return nil, err
	}
	evidence, err := replayedEvidence(ctx, options, art)
	if err != nil {
		return nil, err
	}
	evidence.Verdict = verdict
	return gradeRun(evidence, labels)
}

func gradeRun(evidence score.Evidence, labels []*model.GroundTruthRecord) (*score.RunScore, error) {
	graded, err := score.ScoreRun(evidence, labels)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return graded, nil
}

func Document(graded *score.RunScore) any {
	return score.Document(graded)
}
