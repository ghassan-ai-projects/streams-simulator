package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func replayAdvance(ctx context.Context, r *Run, args map[string]any) error {
	_, err := r.Advance(ctx, commandTime(args, "to_ns"), false)
	return wrapReplayError(err)
}

func replayFault(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpFaultClear {
		return r.ClearFault(commandString(args, "fault_id"), commandTime(args, "at_ns"))
	}
	_, err := r.InjectFault(commandString(args, "entity_id"), commandString(args, "fault"), commandTime(args, "onset_ns"), asMap(args["params"]))
	return wrapReplayError(err)
}

func replayPerturb(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpPerturbClear {
		return r.ClearPerturb(commandString(args, "perturb_id"))
	}
	_, err := r.ApplyPerturb(commandString(args, "perturbation"), asMap(args["params"]), commandTime(args, "from_ns"), commandTime(args, "until_ns"))
	return wrapReplayError(err)
}

func replayActuation(r *Run, cmd *model.Command) error {
	args := cmd.Args
	switch cmd.Op {
	case model.OpEffectorInvoke:
		return replayEffector(r, args)
	case model.OpEntityAdd:
		return r.AddEntity(commandString(args, "entity_id"), commandTime(args, "at_ns"))
	case model.OpEntityRetire:
		return r.RetireEntity(commandString(args, "entity_id"), commandString(args, "reason"), commandTime(args, "at_ns"))
	case model.OpEnvInject, model.OpWorldCreate, model.OpRunBegin, model.OpRunEnd, model.OpClockRun:
		return nil
	}
	return fmt.Errorf("run: unknown command op %q", cmd.Op)
}

func replayEffector(r *Run, args map[string]any) error {
	_, err := r.InvokeEffector(commandString(args, "effector"), commandString(args, "entity_id"), commandString(args, "command_id"), asMap(args["args"]), commandTime(args, "at_ns"))
	return wrapReplayError(err)
}

func wrapReplayError(err error) error {
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}
