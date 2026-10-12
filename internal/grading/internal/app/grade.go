// Package app is the use-case layer of the grading module: it reads the files a
// run published, replays the artifact and scores the evidence.
package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

// Options name the run to grade: its artifact, the sealed label (empty selects
// label.json beside the artifact), and the directories the artifact's domain
// and adapter are looked up in to replay it.
type Options struct{ Artifact, Label, DomainsDir, AdaptersDir string }

// Grade scores a finished run from its artifacts with the same scorer the
// director uses online, so the two scorecards agree. Delivery and state
// evidence come from the files the run published; the effector calls come
// from replaying the artifact, because the command log records the
// invocations but not whether each took effect. A replay that does not
// reproduce the recorded trace is refused rather than graded.
func Grade(ctx context.Context, options Options) (*score.Scorecard, error) {
	art, err := run.LoadArtifact(options.Artifact)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return gradeArtifact(ctx, options, art)
}

// gradeArtifact grades a loaded artifact against the reports beside it.
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
