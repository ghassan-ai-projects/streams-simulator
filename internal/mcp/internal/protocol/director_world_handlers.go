package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
)

func (d directorTools) handleWorldCreate(_ context.Context, args map[string]any) (any, error) {
	return d.CreateWorld(args)
}
func (d directorTools) handleWorldDescribe(_ context.Context, args map[string]any) (any, error) {
	return d.DescribeWorld(app.Str(args, "world_id"))
}
func (d directorTools) handleWorldDestroy(_ context.Context, args map[string]any) (any, error) {
	return d.DestroyWorld(app.Str(args, "world_id"))
}
func (d directorTools) handleClockAdvance(ctx context.Context, args map[string]any) (any, error) {
	worldID := app.Str(args, "world_id")
	toNS, err := d.clockTarget(args, worldID)
	if err != nil {
		return nil, err
	}
	return d.Advance(ctx, worldID, toNS, boolArg(args, "await_consumer"))
}
func (d directorTools) handleClockState(_ context.Context, args map[string]any) (any, error) {
	return d.ClockState(app.Str(args, "world_id"))
}

func (d directorTools) relativeClockTarget(worldID string, byNS int64) (int64, error) {
	if byNS < 0 {
		return 0, app.ToolErrorf(app.CodeInvalidArgs, "by_ns must be non-negative")
	}
	w := d.World(worldID)
	if w == nil {
		return 0, app.ToolErrorf(app.CodeWorldNotFound, "unknown world %q", worldID)
	}
	const maxInt64 = int64(1<<63 - 1)
	if byNS > maxInt64-w.Run.Status().ClockNS {
		return 0, app.ToolErrorf(app.CodeInvalidArgs, "by_ns overflows the world clock")
	}
	return w.Run.Status().ClockNS + byNS, nil
}

func (d directorTools) clockTarget(args map[string]any, worldID string) (int64, error) {
	toNS, hasTo := app.IntArg(args, "to_ns")
	byNS, hasBy := app.IntArg(args, "by_ns")
	if hasTo == hasBy {
		return 0, app.ToolErrorf(app.CodeInvalidArgs, "exactly one of by_ns or to_ns is required")
	}
	if hasBy {
		var err error
		toNS, err = d.relativeClockTarget(worldID, byNS)
		if err != nil {
			return 0, err
		}
	}
	return toNS, nil
}
