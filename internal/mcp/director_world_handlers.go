package mcp

import (
	"context"
)

func (d *Director) handleWorldCreate(_ context.Context, args map[string]any) (any, error) {
	return d.CreateWorld(args)
}
func (d *Director) handleWorldDescribe(_ context.Context, args map[string]any) (any, error) {
	return d.DescribeWorld(str(args, "world_id"))
}
func (d *Director) handleWorldDestroy(_ context.Context, args map[string]any) (any, error) {
	return d.DestroyWorld(str(args, "world_id"))
}
func (d *Director) handleClockAdvance(ctx context.Context, args map[string]any) (any, error) {
	worldID := str(args, "world_id")
	toNS, err := d.clockTarget(args, worldID)
	if err != nil {
		return nil, err
	}
	return d.Advance(ctx, worldID, toNS, boolArg(args, "await_consumer"))
}
func (d *Director) handleClockState(_ context.Context, args map[string]any) (any, error) {
	return d.ClockState(str(args, "world_id"))
}

func (d *Director) relativeClockTarget(worldID string, byNS int64) (int64, error) {
	if byNS < 0 {
		return 0, errTool(CodeInvalidArgs, "by_ns must be non-negative")
	}
	w := d.World(worldID)
	if w == nil {
		return 0, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	const maxInt64 = int64(1<<63 - 1)
	if byNS > maxInt64-w.Run.World.Clock() {
		return 0, errTool(CodeInvalidArgs, "by_ns overflows the world clock")
	}
	return w.Run.World.Clock() + byNS, nil
}

func (d *Director) clockTarget(args map[string]any, worldID string) (int64, error) {
	toNS, hasTo := intArg(args, "to_ns")
	byNS, hasBy := intArg(args, "by_ns")
	if hasTo == hasBy {
		return 0, errTool(CodeInvalidArgs, "exactly one of by_ns or to_ns is required")
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
