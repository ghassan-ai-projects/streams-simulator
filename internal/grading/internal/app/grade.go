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

func gradeArtifact(ctx context.Context, options Options, art *model.RunArtifact) (*score.RunScore, error) {
	verdict, labels, err := loadReports(filepath.Dir(options.Artifact), options.Label)
	if err != nil {
		return nil, err
	}
	evidence, err := verdictEvidence(ctx, options, art, verdict)
	if err != nil {
		return nil, err
	}
	return gradeRun(evidence, labels)
}

func verdictEvidence(ctx context.Context, options Options, art *model.RunArtifact, verdict *model.Verdict) (score.Evidence, error) {
	evidence, err := replayedEvidence(ctx, options, art)
	if err != nil {
		return score.Evidence{}, err
	}
	evidence.Verdict = verdict
	return withDoNothing(ctx, options, art, evidence)
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
