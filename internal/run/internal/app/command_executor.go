package app

import (
	"context"
	"errors"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type refusalPolicy int

const (
	keepEveryRefusal refusalPolicy = iota

	keepWorldRefusals
)

func (policy refusalPolicy) keeps(err error) bool {
	if policy == keepEveryRefusal {
		return true
	}
	return errors.Is(err, world.ErrInterlockRefused) || errors.Is(err, world.ErrCommandIDReused)
}

type StepResult struct {
	Acknowledgement *model.EffectorAcknowledgement
	FaultID         string
}

func executeCommand(ctx context.Context, r *Run, cmd *model.Command, policy refusalPolicy) error {
	_, err := executeObserved(ctx, r, cmd, policy)
	return err
}

func executeObserved(ctx context.Context, r *Run, cmd *model.Command, policy refusalPolicy) (StepResult, error) {
	switch cmd.Op {
	case model.OpEffectorInvoke:
		ack, err := applyInvocation(r, cmd.Args, policy)
		return StepResult{Acknowledgement: ack}, err
	case model.OpFaultInject:
		id, err := applyInjection(r, cmd.Args)
		return StepResult{FaultID: id}, err
	}
	return StepResult{}, executeOther(ctx, r, cmd)
}

func executeOther(ctx context.Context, r *Run, cmd *model.Command) error {
	switch cmd.Op {
	case model.OpClockAdvance:
		return applyAdvance(ctx, r, cmd.Args)
	case model.OpFaultClear:
		return r.ClearFault(rules.CommandString(cmd.Args, "fault_id"), rules.CommandTime(cmd.Args, "at_ns"))
	case model.OpPerturbApply, model.OpPerturbClear:
		return applyPerturbation(r, cmd)
	default:
		return applyLifecycle(r, cmd)
	}
}

func applyAdvance(ctx context.Context, r *Run, args map[string]any) error {
	_, err := r.Advance(ctx, rules.CommandTime(args, "to_ns"), false)
	return wrapExecutionError(err)
}

func applyInjection(r *Run, args map[string]any) (string, error) {
	id, err := r.InjectFault(rules.CommandString(args, "entity_id"), rules.CommandString(args, "fault"), rules.CommandTime(args, "onset_ns"), rules.AsMap(args["params"]))
	return id, wrapExecutionError(err)
}

func applyPerturbation(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpPerturbClear {
		return r.ClearPerturb(rules.CommandString(args, "perturb_id"))
	}
	_, err := r.ApplyPerturb(rules.CommandString(args, "perturbation"), rules.AsMap(args["params"]), rules.CommandTime(args, "from_ns"), rules.CommandTime(args, "until_ns"))
	return wrapExecutionError(err)
}

func applyInvocation(r *Run, args map[string]any, policy refusalPolicy) (*model.EffectorAcknowledgement, error) {
	commandID := rules.CommandString(args, "command_id")
	result, err := r.InvokeEffector(rules.CommandString(args, "effector"), rules.CommandString(args, "entity_id"), commandID, rules.AsMap(args["args"]), rules.CommandTime(args, "at_ns"))
	return acknowledgement(r, commandID, result, err, policy)
}

func acknowledgement(r *Run, commandID string, result *world.InvokeResult, err error, policy refusalPolicy) (*model.EffectorAcknowledgement, error) {
	if err == nil || policy.keeps(err) {
		return invocationAcknowledgement(r, commandID, result, err), nil
	}
	return nil, wrapExecutionError(err)
}

func invocationAcknowledgement(r *Run, commandID string, result *world.InvokeResult, err error) *model.EffectorAcknowledgement {
	switch {
	case err == nil:
		return world.Acknowledge(result)
	case errors.Is(err, world.ErrInterlockRefused):
		return refusedAcknowledgement(r, commandID, "interlock_refused")
	case errors.Is(err, world.ErrCommandIDReused):
		return refusedAcknowledgement(r, commandID, "command_id_reused")
	}
	return nil
}

func refusedAcknowledgement(r *Run, commandID, reason string) *model.EffectorAcknowledgement {
	return &model.EffectorAcknowledgement{Simulated: true, WorldID: r.World.ID, CommandID: commandID, Reason: reason}
}

func applyLifecycle(r *Run, cmd *model.Command) error {
	args := cmd.Args
	switch cmd.Op {
	case model.OpEntityAdd:
		return r.AddEntity(rules.CommandString(args, "entity_id"), rules.CommandTime(args, "at_ns"))
	case model.OpEntityRetire:
		return r.RetireEntity(rules.CommandString(args, "entity_id"), rules.CommandString(args, "reason"), rules.CommandTime(args, "at_ns"))
	case model.OpEnvInject, model.OpWorldCreate, model.OpRunBegin, model.OpRunEnd, model.OpClockRun:
		return nil
	}
	return fmt.Errorf("run: unknown command op %q", cmd.Op)
}

func wrapExecutionError(err error) error {
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}
