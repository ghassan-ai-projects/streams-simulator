package app

import (
	"context"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func BuildVerified(ctx context.Context, dir string, dirs run.Directories) (*domain.Report, error) {
	built, err := Build(ctx, dir)
	if err != nil || !built.Run.Ended {
		return built, err
	}
	return domain.WithReplay(built, replayOf(ctx, dir, dirs)), nil
}

func replayOf(ctx context.Context, dir string, dirs run.Directories) domain.Replay {
	result, err := run.ReplayFile(ctx, filepath.Join(dir, model.FileRunArtifact), dirs)
	switch {
	case err != nil:
		return domain.Replay{Status: domain.NotChecked, Detail: "The replay could not run: " + insideFolder(err, dir)}
	case result.Matches:
		return domain.Replay{Status: domain.ReplayReproduces}
	case result.Incomplete:
		return domain.Replay{Status: domain.ReplayIncomplete, Detail: withoutFolder(result.Detail, dir)}
	}
	return domain.Replay{Status: domain.ReplayDiverges, Detail: withoutFolder(result.Detail, dir)}
}
