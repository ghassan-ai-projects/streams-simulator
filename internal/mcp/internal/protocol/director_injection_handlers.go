package protocol

import (
	"context"
)

func (d directorTools) handleFaultInject(_ context.Context, args map[string]any) (any, error) {
	return d.InjectFault(stringArg(args, "world_id"), stringArg(args, "entity_id"), stringArg(args, "fault"), numberArg(args, "onset_ns", 0), mapArg(args, "params"))
}
func (d directorTools) handleFaultClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearFault(stringArg(args, "world_id"), stringArg(args, "fault_id"))
}
func (d directorTools) handleFaultList(_ context.Context, args map[string]any) (any, error) {
	return d.ListFaults(stringArg(args, "world_id"))
}
func (d directorTools) handlePerturbApply(_ context.Context, args map[string]any) (any, error) {
	return d.ApplyPerturb(stringArg(args, "world_id"), stringArg(args, "perturbation"), mapArg(args, "params"), numberArg(args, "from_ns", 0), numberArg(args, "until_ns", 0))
}
func (d directorTools) handlePerturbClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearPerturb(stringArg(args, "world_id"), stringArg(args, "perturb_id"))
}
func (d directorTools) handleEnvInject(_ context.Context, args map[string]any) (any, error) {
	return d.EnvInject(stringArg(args, "world_id"), stringArg(args, "target"), stringArg(args, "fault"), mapArg(args, "params"))
}
