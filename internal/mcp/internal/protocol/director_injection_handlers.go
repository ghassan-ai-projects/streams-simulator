package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
)

func (d directorTools) handleFaultInject(_ context.Context, args map[string]any) (any, error) {
	return d.InjectFault(app.Str(args, "world_id"), app.Str(args, "entity_id"), app.Str(args, "fault"), app.Num(args, "onset_ns", 0), mapArg(args, "params"))
}
func (d directorTools) handleFaultClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearFault(app.Str(args, "world_id"), app.Str(args, "fault_id"))
}
func (d directorTools) handleFaultList(_ context.Context, args map[string]any) (any, error) {
	return d.ListFaults(app.Str(args, "world_id"))
}
func (d directorTools) handlePerturbApply(_ context.Context, args map[string]any) (any, error) {
	return d.ApplyPerturb(app.Str(args, "world_id"), app.Str(args, "perturbation"), mapArg(args, "params"), app.Num(args, "from_ns", 0), app.Num(args, "until_ns", 0))
}
func (d directorTools) handlePerturbClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearPerturb(app.Str(args, "world_id"), app.Str(args, "perturb_id"))
}
func (d directorTools) handleEnvInject(_ context.Context, args map[string]any) (any, error) {
	return d.EnvInject(app.Str(args, "world_id"), app.Str(args, "target"), app.Str(args, "fault"), mapArg(args, "params"))
}
