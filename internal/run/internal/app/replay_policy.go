package app

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
)

func playingStep(play *rules.PolicyPlay) commandStep {
	return func(ctx context.Context, r *Run, cmd *model.Command, policy refusalPolicy) error {
		if cmd.Op != model.OpClockAdvance {
			return executeCommand(ctx, r, cmd, policy)
		}
		return playThrough(ctx, r, play, rules.CommandTime(cmd.Args, "to_ns"))
	}
}

func playThrough(ctx context.Context, r *Run, play *rules.PolicyPlay, toNS int64) error {
	for stop, ok := play.NextStop(toNS); ok; stop, ok = play.NextStop(toNS) {
		if err := advancePlay(ctx, r, stop); err != nil {
			return err
		}
		if err := playAt(r, play, stop); err != nil {
			return err
		}
	}
	return advancePlay(ctx, r, toNS)
}

func advancePlay(ctx context.Context, r *Run, atNS int64) error {
	if atNS <= r.World.Clock() {
		return nil
	}
	_, err := r.Advance(ctx, atNS, false)
	return wrapOutcome(err)
}

func playAt(r *Run, play *rules.PolicyPlay, atNS int64) error {
	if play.SampleDue(atNS) {
		play.Observe(atNS, conditionByInstance(r, play, atNS))
	}
	for _, call := range play.Due(atNS) {
		if err := invokePolicyCall(r, call); err != nil {
			return err
		}
	}
	return nil
}

func conditionByInstance(r *Run, play *rules.PolicyPlay, atNS int64) map[string]bool {
	holding, condition := map[string]bool{}, play.Condition()
	for _, id := range r.World.EntityIDs() {
		if r.World.Entity(id).Type == play.Kind() {
			holding[id] = r.World.ConditionHolds(id, condition, atNS)
		}
	}
	return holding
}

func invokePolicyCall(r *Run, call rules.PolicyCall) error {
	result, invoked := r.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS)
	_, err := acknowledgement(r, call.CommandID, result, invoked, keepWorldRefusals)
	return wrapOutcome(err)
}
