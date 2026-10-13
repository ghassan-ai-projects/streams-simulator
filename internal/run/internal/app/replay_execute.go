package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func newReplayRun(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*Run, error) {
	cfg := replayConfig(art, spec, adapterSpec, sinkTarget)
	replay, err := New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if art.WorldDigest != "" && replay.Digest() != art.WorldDigest {
		return nil, &digestMismatch{"world", art.WorldDigest, replay.Digest()}
	}
	if art.Unblinded && art.Mode != journal.ModeExplore {
		replay.Unblind()
	}
	return replay, nil
}

func replayCommands(ctx context.Context, replay *Run, commands []model.Command) error {
	for _, command := range commands {
		if err := executeCommand(ctx, replay, &command, keepEveryRefusal); err != nil {
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
	traceMatches := result.GotDigest == result.WantDigest
	if !traceMatches {
		noteReplayDivergence(replay, art, result)
	}
	result.Matches = noteIdentityDivergence(replay, art, result) && traceMatches
	return result
}

func noteIdentityDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) bool {
	noteJournalDivergence(replay, art, result)
	stampMatches := noteStampDivergence(replay, art, result)
	return stampMatches && len(result.JournalDiverged) == 0
}

func noteStampDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) bool {
	if replay.unblinded == art.Unblinded {
		return true
	}
	result.Detail = divergenceDetail(result.Detail,
		fmt.Sprintf("the artifact says unblinded=%v, the replay of a %s run is unblinded=%v", art.Unblinded, replay.mode(), replay.unblinded))
	return false
}

func noteReplayDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	index := firstDivergentRecord(replay, art)
	if index >= 0 {
		result.FirstDivergence = &index
		result.Detail = divergenceDetail(result.Detail,
			fmt.Sprintf("replayed %d records, artifact recorded %d", len(replay.ledger), art.Counts.Emitted))
		return
	}
	result.Detail = divergenceDetail(result.Detail,
		"trace digest differs with the same record count; the artifact keeps only the digest, so the first differing record is not known")
}

func divergenceDetail(existing, note string) string {
	if existing == "" {
		return note
	}
	return existing + "; " + note
}
