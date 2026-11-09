package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/files"
)

func scoreOf(ctx context.Context, dir string, folder domain.Folder) (*domain.Scoring, error) {
	if folder.Artifact == nil || folder.Verdict == nil || folder.Mismatch != "" {
		return nil, nil
	}
	labeled, err := holdsLabels(dir)
	if err != nil || !labeled {
		return nil, err
	}
	return grade(ctx, filepath.Join(dir, model.FileRunArtifact))
}

func holdsLabels(dir string) (bool, error) {
	for _, name := range []string{model.FileLabel, model.FileLabels} {
		label, err := files.Read(filepath.Join(dir, name))
		if err != nil || label.Present {
			return label.Present, err
		}
	}
	return false, nil
}

func grade(ctx context.Context, artifact string) (*domain.Scoring, error) {
	graded, err := grading.GradeRun(ctx, grading.Options{Artifact: artifact})
	if errors.Is(err, grading.ErrRunUnblinded) {
		return &domain.Scoring{Status: domain.NotScored, Reason: err.Error()}, nil
	}
	if err != nil {
		return &domain.Scoring{Status: domain.Failed, Reason: insideFolder(err, filepath.Dir(artifact))}, nil
	}
	raw, err := json.Marshal(grading.Document(graded))
	if err != nil {
		return nil, fmt.Errorf("report: encode scorecard: %w", err)
	}
	return &domain.Scoring{Status: domain.Scored, Scorecard: raw}, nil
}

func withOutcome(ctx context.Context, dir string, built *domain.Report) (*domain.Report, error) {
	artifact, err := loadArtifact(dir)
	if err != nil {
		return nil, err
	}
	if built.TruthUnavailable == "" && domain.DeclaresOutcome(artifact) {
		built.OutcomeScore = outcomeScoreOf(ctx, dir)
	}
	return built, nil
}

func outcomeScoreOf(ctx context.Context, dir string) *domain.OutcomeScoring {
	graded, err := grading.GradeOutcome(ctx, grading.OutcomeOptions{RunDir: dir})
	switch {
	case errors.Is(err, grading.ErrRunUnblinded):
		return &domain.OutcomeScoring{Status: domain.NotScored, Reason: err.Error()}
	case err != nil:
		return &domain.OutcomeScoring{Status: domain.Failed, Reason: insideFolder(err, dir)}
	}
	return &domain.OutcomeScoring{Status: domain.Scored, Result: graded}
}

func insideFolder(err error, dir string) string {
	return withoutFolder(err.Error(), dir)
}

func withoutFolder(text, dir string) string {
	return strings.ReplaceAll(text, filepath.Clean(dir)+string(filepath.Separator), "")
}
