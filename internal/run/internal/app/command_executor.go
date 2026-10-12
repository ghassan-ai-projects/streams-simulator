package app

import (
	"context"
	"errors"
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// refusalPolicy says which effector refusals a command list keeps as recorded
// outcomes instead of failing on them.
type refusalPolicy int

const (
	// keepEveryRefusal replays a recorded command log. The log holds every
	// call the world refused, and the original run got the same refusal, so
	// each one is history being reproduced; a divergence shows in the digests.
	keepEveryRefusal refusalPolicy = iota
	// keepWorldRefusals executes authored commands. Only the refusals the
	// world decides as it runs (an interlock, a reused command id) are
	// outcomes; any other refusal means the command itself was wrong.
	keepWorldRefusals
)

// keeps reports whether an effector refusal is a recorded outcome.
func (policy refusalPolicy) keeps(err error) bool {
	if policy == keepEveryRefusal {
		return true
	}
	return errors.Is(err, world.ErrInterlockRefused) || errors.Is(err, world.ErrCommandIDReused)
}

// executeCommand applies one command, recorded or authored, to a run: the
// single executor behind replay and scenario runs.
func executeCommand(ctx context.Context, r *Run, cmd *model.Command, policy refusalPolicy) error {
	switch cmd.Op {
	case model.OpClockAdvance:
		return applyAdvance(ctx, r, cmd.Args)
	case model.OpFaultInject, model.OpFaultClear:
		return applyFault(r, cmd)
	case model.OpPerturbApply, model.OpPerturbClear:
		return applyPerturbation(r, cmd)
	case model.OpEffectorInvoke:
		return applyInvocation(r, cmd.Args, policy)
	default:
		return applyLifecycle(r, cmd)
	}
}

// applyAdvance never awaits a consumer: whether a live session waited for its
// consumer is pacing, and the consumer's own commands are in the list.
func applyAdvance(ctx context.Context, r *Run, args map[string]any) error {
	_, err := r.Advance(ctx, rules.CommandTime(args, "to_ns"), false)
	return wrapExecutionError(err)
}

func applyFault(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpFaultClear {
		return r.ClearFault(rules.CommandString(args, "fault_id"), rules.CommandTime(args, "at_ns"))
	}
	_, err := r.InjectFault(rules.CommandString(args, "entity_id"), rules.CommandString(args, "fault"), rules.CommandTime(args, "onset_ns"), rules.AsMap(args["params"]))
	return wrapExecutionError(err)
}

func applyPerturbation(r *Run, cmd *model.Command) error {
	args := cmd.Args
	if cmd.Op == model.OpPerturbClear {
		return r.ClearPerturb(rules.CommandString(args, "perturb_id"))
	}
	_, err := r.ApplyPerturb(rules.CommandString(args, "perturbation"), rules.AsMap(args["params"]), rules.CommandTime(args, "from_ns"), rules.CommandTime(args, "until_ns"))
	return wrapExecutionError(err)
}

// applyInvocation invokes an effector; the policy decides whether a refusal
// is an outcome or a failure. The run records the call either way.
func applyInvocation(r *Run, args map[string]any, policy refusalPolicy) error {
	_, err := r.InvokeEffector(rules.CommandString(args, "effector"), rules.CommandString(args, "entity_id"), rules.CommandString(args, "command_id"), rules.AsMap(args["args"]), rules.CommandTime(args, "at_ns"))
	if err == nil || policy.keeps(err) {
		return nil
	}
	return wrapExecutionError(err)
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
