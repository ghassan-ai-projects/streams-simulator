package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func newReplayRun(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*Run, error) {
	return startReplay(ctx, art, replayConfig(art, spec, adapterSpec, sinkTarget))
}

func startReplay(ctx context.Context, art *model.RunArtifact, cfg Config) (*Run, error) {
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
	return replayEach(ctx, replay, commands, executeCommand)
}

type commandStep func(context.Context, *Run, *model.Command, refusalPolicy) error

func replayEach(ctx context.Context, replay *Run, commands []model.Command, execute commandStep) error {
	for _, command := range commands {
		if err := execute(ctx, replay, &command, keepEveryRefusal); err != nil {
			return fmt.Errorf("run: replay command %d (%s): %w", command.Seq, command.Op, err)
		}
	}
	if _, err := replay.End(""); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func finishReplayResult(replay *Run, art *model.RunArtifact, result *ReplayResult) *ReplayResult {
	recordReplayOutcome(replay, art, result)
	result.Matches, result.Reason = true, ReasonMatch
	noteMissingDigests(art, result)
	noteTraceDivergence(replay, art, result)
	noteJournalDivergence(replay, art, result)
	noteEvidenceDivergence(replay, art, result)
	noteStampDivergence(replay, art, result)
	return result
}

func recordReplayOutcome(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	result.GotDigest = replay.TraceDigest()
	result.Emitted = replay.World.EmittedCount()
	result.Incomplete = art.Incomplete
	if art.Incomplete && result.Detail == "" {
		result.Detail = art.Error
	}
}

func noteStampDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	if replay.unblinded == art.Unblinded {
		return
	}
	diverge(result, ReasonUnblindedStamp, "",
		fmt.Sprintf("the artifact says unblinded=%v, the replay of a %s run is unblinded=%v", art.Unblinded, replay.mode(), replay.unblinded))
}

func noteTraceDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	if result.GotDigest == result.WantDigest || result.WantDigest == "" {
		return
	}
	index := firstDivergentRecord(replay, art)
	if index < 0 {
		diverge(result, ReasonTraceDivergence, "",
			"trace digest differs with the same record count; the artifact keeps only the digest, so the first differing record is not known")
		return
	}
	result.FirstDivergence = &index
	diverge(result, ReasonTraceDivergence, "", fmt.Sprintf("replayed %d records, artifact recorded %d", len(replay.ledger), art.Counts.Emitted))
}
