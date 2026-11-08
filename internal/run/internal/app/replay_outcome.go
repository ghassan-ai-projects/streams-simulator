package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
)

var (
	ErrUnknownOutcomeMode = rules.ErrUnknownOutcomeMode
	ErrUnknownPolicy      = rules.ErrUnknownPolicy
)

func ReplayOutcome(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, mode model.OutcomeMode) (*model.OutcomeFacts, error) {
	spec, adapterSpec, _, err := prepareReplay(art, spec, adapterSpec)
	if err != nil {
		return nil, err
	}
	commands, err := rules.OutcomeCommands(art.CommandLog, mode)
	if err != nil {
		return nil, err
	}
	step, err := outcomeStep(spec, mode, art.WorldConfig.StartTimeNS)
	if err != nil {
		return nil, err
	}
	return replayOutcomeRun(ctx, art, outcomeConfig(art, spec, adapterSpec), commands, step)
}

func outcomeConfig(art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) Config {
	cfg := replayConfig(art, spec, adapterSpec, "")
	cfg.silent, cfg.recordOutcome, cfg.retainJournal = true, true, false
	return cfg
}

func outcomeStep(spec *domain.Compiled, mode model.OutcomeMode, startNS int64) (commandStep, error) {
	name, isPolicy := mode.Policy()
	if !isPolicy {
		return executeCommand, nil
	}
	play, err := rules.NewPolicyPlay(spec.Spec, name, startNS)
	if err != nil {
		return nil, err
	}
	return playingStep(play), nil
}

func replayOutcomeRun(ctx context.Context, art *model.RunArtifact, cfg Config, commands []model.Command, step commandStep) (*model.OutcomeFacts, error) {
	replay, err := startReplay(ctx, art, cfg)
	if err != nil {
		return nil, err
	}
	if err := replayEach(ctx, replay, commands, step); err != nil {
		return nil, err
	}
	facts := replay.World.OutcomeFacts(replay.World.Clock())
	facts.Calls = outcomeCalls(replay)
	return facts, nil
}

func outcomeCalls(replay *Run) []model.OutcomeCall {
	calls := []model.OutcomeCall{}
	for _, call := range replay.World.EffectorCalls() {
		calls = append(calls, model.OutcomeCall{Effector: call.Effector, EntityID: call.EntityID, AtNS: call.AtNS, Accepted: call.Accepted})
	}
	return calls
}

func wrapOutcome(err error) error {
	if err != nil {
		return fmt.Errorf("run: outcome replay: %w", err)
	}
	return nil
}
