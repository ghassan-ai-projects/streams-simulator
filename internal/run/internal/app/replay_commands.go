package app

import (
	"context"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func replayAdvance(ctx context.Context, r *Run, args map[string]any) error {
	_, err := r.Advance(ctx, rules.CommandTime(args, "to_ns"), false)
	return wrapReplayError(err)
}

func replayFault(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpFaultClear {
		return r.ClearFault(rules.CommandString(args, "fault_id"), rules.CommandTime(args, "at_ns"))
	}
	_, err := r.InjectFault(rules.CommandString(args, "entity_id"), rules.CommandString(args, "fault"), rules.CommandTime(args, "onset_ns"), rules.AsMap(args["params"]))
	return wrapReplayError(err)
}

func replayPerturb(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpPerturbClear {
		return r.ClearPerturb(rules.CommandString(args, "perturb_id"))
	}
	_, err := r.ApplyPerturb(rules.CommandString(args, "perturbation"), rules.AsMap(args["params"]), rules.CommandTime(args, "from_ns"), rules.CommandTime(args, "until_ns"))
	return wrapReplayError(err)
}

func replayActuation(r *Run, cmd *model.Command) error {
	args := cmd.Args
	switch cmd.Op {
	case model.OpEffectorInvoke:
		return replayEffector(r, args)
	case model.OpEntityAdd:
		return r.AddEntity(rules.CommandString(args, "entity_id"), rules.CommandTime(args, "at_ns"))
	case model.OpEntityRetire:
		return r.RetireEntity(rules.CommandString(args, "entity_id"), rules.CommandString(args, "reason"), rules.CommandTime(args, "at_ns"))
	case model.OpEnvInject, model.OpWorldCreate, model.OpRunBegin, model.OpRunEnd, model.OpClockRun:
		return nil
	}
	return fmt.Errorf("run: unknown command op %q", cmd.Op)
}

// replayEffector repeats a logged invocation. The log also holds the
// invocations the world refused; the original run got the same refusal, so it
// is part of the history being reproduced, not a failure of the replay. A
// divergence still shows in the digests.
func replayEffector(r *Run, args map[string]any) error {
	_, _ = r.InvokeEffector(rules.CommandString(args, "effector"), rules.CommandString(args, "entity_id"), rules.CommandString(args, "command_id"), rules.AsMap(args["args"]), rules.CommandTime(args, "at_ns"))
	return nil
}

func wrapReplayError(err error) error {
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}
