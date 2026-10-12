package protocol

import (
	"context"
)

func (d directorTools) handleRunBegin(_ context.Context, args map[string]any) (any, error) {
	return d.BeginRun(stringArg(args, "world_id"), stringArg(args, "label"))
}
func (d directorTools) handleRunEnd(_ context.Context, args map[string]any) (any, error) {
	return d.EndRun(stringArg(args, "world_id"))
}
func (d directorTools) handleRunVerify(_ context.Context, args map[string]any) (any, error) {
	return d.VerifyRun(stringArg(args, "run_artifact_path"))
}
