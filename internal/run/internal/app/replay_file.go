package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Directories struct{ Domains, Adapters string }

func ReplayFile(ctx context.Context, artifactPath string, dirs Directories) (*ReplayResult, error) {
	art, err := LoadArtifact(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	spec, adap, err := LoadInputs(dirs.Domains, dirs.Adapters, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return replayWithJournalFiles(ctx, art, spec, adap, filepath.Dir(artifactPath))
}

func replayWithJournalFiles(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter, dir string) (*ReplayResult, error) {
	res, replay, err := replayArtifactRun(ctx, art, spec, adap, "")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := errors.Join(checkJournalFiles(dir, replay, art, res), checkEvidenceFiles(dir, replay, art, res)); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return res, nil
}
