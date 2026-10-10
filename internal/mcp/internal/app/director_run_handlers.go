package app

import (
	"context"
)

func (d *Director) handleRunBegin(_ context.Context, args map[string]any) (any, error) {
	return d.BeginRun(str(args, "world_id"), str(args, "label"))
}
func (d *Director) handleRunEnd(_ context.Context, args map[string]any) (any, error) {
	return d.EndRun(str(args, "world_id"))
}
func (d *Director) handleRunVerify(_ context.Context, args map[string]any) (any, error) {
	return d.VerifyRun(str(args, "run_artifact_path"))
}
