package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Directories say where the domains and adapters a run names are looked up.
type Directories struct{ Domains, Adapters string }

// ReplayFile reproduces the run artifact at artifactPath against the domain
// and adapter it recorded, found in the directories, and reports whether the
// trace digest matches.
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

// replayWithJournalFiles replays the artifact and checks the journal files in
// dir against the journals the replay rebuilt.
func replayWithJournalFiles(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter, dir string) (*ReplayResult, error) {
	res, replay, err := replayArtifactRun(ctx, art, spec, adap, "")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := checkJournalFiles(dir, replay, art, res); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return res, nil
}
