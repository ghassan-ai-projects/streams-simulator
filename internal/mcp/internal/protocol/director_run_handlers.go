package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
)

func (d directorTools) handleRunBegin(_ context.Context, args map[string]any) (any, error) {
	return d.BeginRun(app.Str(args, "world_id"), app.Str(args, "label"))
}
func (d directorTools) handleRunEnd(_ context.Context, args map[string]any) (any, error) {
	return d.EndRun(app.Str(args, "world_id"))
}
func (d directorTools) handleRunVerify(_ context.Context, args map[string]any) (any, error) {
	return d.VerifyRun(app.Str(args, "run_artifact_path"))
}
