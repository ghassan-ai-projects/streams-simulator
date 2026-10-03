package mcp

import (
	"context"
)

func (d *Director) handleFaultInject(_ context.Context, args map[string]any) (any, error) {
	return d.InjectFault(str(args, "world_id"), str(args, "entity_id"), str(args, "fault"), num(args, "onset_ns", 0), mapArg(args, "params"))
}
func (d *Director) handleFaultClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearFault(str(args, "world_id"), str(args, "fault_id"))
}
func (d *Director) handleFaultList(_ context.Context, args map[string]any) (any, error) {
	return d.ListFaults(str(args, "world_id"))
}
func (d *Director) handlePerturbApply(_ context.Context, args map[string]any) (any, error) {
	return d.ApplyPerturb(str(args, "world_id"), str(args, "perturbation"), mapArg(args, "params"), num(args, "from_ns", 0), num(args, "until_ns", 0))
}
func (d *Director) handlePerturbClear(_ context.Context, args map[string]any) (any, error) {
	return d.ClearPerturb(str(args, "world_id"), str(args, "perturb_id"))
}
func (d *Director) handleEnvInject(_ context.Context, args map[string]any) (any, error) {
	return d.EnvInject(str(args, "world_id"), str(args, "target"), str(args, "fault"), mapArg(args, "params"))
}
