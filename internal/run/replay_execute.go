package run

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func newReplayRun(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*Run, error) {
	cfg := replayConfig(art, spec, adapterSpec, sinkTarget)
	replay, err := New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if art.WorldDigest != "" && replay.Digest() != art.WorldDigest {
		return nil, fmt.Errorf("run: world digest mismatch: artifact=%s current=%s", art.WorldDigest, replay.Digest())
	}
	return replay, nil
}

func replayCommands(ctx context.Context, replay *Run, commands []model.Command) error {
	for _, command := range commands {
		if err := executeCommand(ctx, replay, &command); err != nil {
			return fmt.Errorf("run: replay command %d (%s): %w", command.Seq, command.Op, err)
		}
	}
	if _, err := replay.End(""); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func finishReplayResult(replay *Run, art *model.RunArtifact, result *ReplayResult) *ReplayResult {
	result.GotDigest = replay.TraceDigest()
	result.Emitted = replay.World.EmittedCount()
	result.Incomplete = art.Incomplete
	if art.Incomplete && result.Detail == "" {
		result.Detail = art.Error
	}
	if result.GotDigest != result.WantDigest {
		noteReplayDivergence(replay, art, result)
		return result
	}
	result.Matches = true
	return result
}

func noteReplayDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	index := firstDivergentRecord(replay, art)
	if index >= 0 {
		result.FirstDivergence = &index
	}
}
