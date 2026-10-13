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
	label, err := files.Read(filepath.Join(dir, model.FileLabel))
	if err != nil || !label.Present {
		return nil, err
	}
	return grade(ctx, filepath.Join(dir, model.FileRunArtifact))
}

func grade(ctx context.Context, artifact string) (*domain.Scoring, error) {
	card, err := grading.Grade(ctx, grading.Options{Artifact: artifact})
	if errors.Is(err, grading.ErrRunUnblinded) {
		return &domain.Scoring{Status: domain.NotScored, Reason: err.Error()}, nil
	}
	if err != nil {
		return &domain.Scoring{Status: domain.Failed, Reason: insideFolder(err, filepath.Dir(artifact))}, nil
	}
	raw, err := json.Marshal(card)
	if err != nil {
		return nil, fmt.Errorf("report: encode scorecard: %w", err)
	}
	return &domain.Scoring{Status: domain.Scored, Scorecard: raw}, nil
}

func insideFolder(err error, dir string) string {
	return strings.ReplaceAll(err.Error(), filepath.Clean(dir)+string(filepath.Separator), "")
}
