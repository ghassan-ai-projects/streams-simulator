package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
)

func (d directorTools) handleWorldCreate(_ context.Context, args map[string]any) (any, error) {
	return d.CreateWorld(worldRequest(args))
}
func (d directorTools) handleWorldDescribe(_ context.Context, args map[string]any) (any, error) {
	return d.DescribeWorld(stringArg(args, "world_id"))
}
func (d directorTools) handleWorldDestroy(_ context.Context, args map[string]any) (any, error) {
	return d.DestroyWorld(stringArg(args, "world_id"))
}
func (d directorTools) handleWorldEffectors(_ context.Context, args map[string]any) (any, error) {
	return d.WorldEffectors(stringArg(args, "world_id"))
}
func (d directorTools) handleWorldInvoke(_ context.Context, args map[string]any) (any, error) {
	return d.InvokeWorldEffector(stringArg(args, "world_id"), stringArg(args, "effector"), stringArg(args, "entity_id"), stringArg(args, "command_id"), mapArg(args, "args"))
}
func (d directorTools) handleClockAdvance(ctx context.Context, args map[string]any) (any, error) {
	worldID := stringArg(args, "world_id")
	toNS, hasTo := intArg(args, "to_ns")
	byNS, hasBy := intArg(args, "by_ns")
	if hasTo == hasBy {
		return nil, director.ToolErrorf(director.CodeInvalidArgs, "exactly one of by_ns or to_ns is required")
	}
	if hasBy {
		return d.AdvanceBy(ctx, worldID, byNS, boolArg(args, "await_consumer"))
	}
	return d.Advance(ctx, worldID, toNS, boolArg(args, "await_consumer"))
}
func (d directorTools) handleClockState(_ context.Context, args map[string]any) (any, error) {
	return d.ClockState(stringArg(args, "world_id"))
}
