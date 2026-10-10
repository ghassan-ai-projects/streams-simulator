package app

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// InjectFault records and applies a world fault.
func (r *Run) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("InjectFault"); err != nil {
		return "", err
	}
	fid, err := r.World.InjectFault(entityID, faultID, onsetNS, params)
	if err != nil {
		return "", fmt.Errorf("run: inject fault: %w", err)
	}
	r.recordWorldCommand(model.OpFaultInject, map[string]any{"entity_id": entityID, "fault": faultID, "onset_ns": onsetNS, "params": params})
	return fid, nil
}

// ClearFault records and clears a fault.
func (r *Run) ClearFault(faultID string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("ClearFault"); err != nil {
		return err
	}

	if err := r.World.ClearFault(faultID, atNS); err != nil {
		return fmt.Errorf("ClearFault: %w", err)
	}
	r.recordWorldCommand(model.OpFaultClear, map[string]any{"fault_id": faultID, "at_ns": atNS})
	return nil
}

// ApplyPerturb records and applies a delivery perturbation.
func (r *Run) ApplyPerturb(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("ApplyPerturb"); err != nil {
		return "", err
	}
	id, err := r.Perturb.Apply(name, params, fromNS, untilNS)
	if err != nil {
		return "", fmt.Errorf("run: perturb: %w", err)
	}
	r.perturbHistory = append(r.perturbHistory, name)
	r.recordWorldCommand(model.OpPerturbApply, map[string]any{"perturbation": name, "params": params, "from_ns": fromNS, "until_ns": untilNS})
	return id, nil
}

// ClearPerturb records and deactivates a perturbation.
func (r *Run) ClearPerturb(id string) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("ClearPerturb"); err != nil {
		return err
	}

	if err := r.Perturb.Clear(id); err != nil {
		return fmt.Errorf("ClearPerturb: %w", err)
	}
	r.recordWorldCommand(model.OpPerturbClear, map[string]any{"perturb_id": id})
	return nil
}

// InvokeEffector records and performs an effector call.
func (r *Run) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("InvokeEffector"); err != nil {
		return nil, err
	}
	result, err := r.World.InvokeEffector(effector, entityID, commandID, args, atNS)
	// Record refusals as well as successful calls so replay reproduces both.
	r.recordWorldCommand(model.OpEffectorInvoke, map[string]any{"effector": effector, "entity_id": entityID, "command_id": commandID, "args": args, "at_ns": atNS})
	if err != nil {
		return nil, fmt.Errorf("InvokeEffector: %w", err)
	}
	return result, nil
}

// AddEntity records and performs an entity birth.
func (r *Run) AddEntity(id string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("AddEntity"); err != nil {
		return err
	}

	if err := r.World.AddEntity(id, atNS); err != nil {
		return fmt.Errorf("AddEntity: %w", err)
	}
	r.recordCommandAt(model.OpEntityAdd, atNS, map[string]any{"entity_id": id, "at_ns": atNS})
	return nil
}

// RetireEntity records and performs an entity retirement.
func (r *Run) RetireEntity(entityID, reason string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("RetireEntity"); err != nil {
		return err
	}

	r.World.Retire(entityID, reason, atNS)
	r.recordCommandAt(model.OpEntityRetire, atNS, map[string]any{"entity_id": entityID, "reason": reason, "at_ns": atNS})
	return nil
}

// ConfigureEnvTarget records an env.inject target (pause/kill/partition).
func (r *Run) ConfigureEnvTarget(target string, allow bool) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.envTargets[target] = target
	r.allowEnv = r.allowEnv || allow
}

// EnvInject records an environment fault against a configured target. No
// environment-fault parameters are declared yet, so any params are rejected
// rather than recorded and ignored.
func (r *Run) EnvInject(target, fault string, params map[string]any, atNS int64) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.refuseWhenFinished("EnvInject"); err != nil {
		return "", err
	}
	if !r.allowEnv {
		return "", fmt.Errorf("run: env.inject not enabled for this world (no configured target)")
	}
	if len(params) > 0 {
		return "", fmt.Errorf("run: env fault %q accepts no parameters (got %d)", fault, len(params))
	}
	r.recordCommandAt(model.OpEnvInject, atNS, map[string]any{"target": target, "fault": fault, "params": params, "at_ns": atNS})
	return "env-" + strconv.Itoa(len(r.commandLog)), nil
}

func (r *Run) recordWorldCommand(op string, args map[string]any) {
	r.recordCommandAt(op, r.World.Clock(), args)
}

func (r *Run) recordCommandAt(op string, atNS int64, args map[string]any) {
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: atNS, Op: op, Args: args,
	})
}

// refuseWhenFinished is the one place a mutating command learns the run is
// over: a finished run's command log, world and evidence are sealed.
func (r *Run) refuseWhenFinished(command string) error {
	if r.finished {
		return fmt.Errorf("%s: run is finished", command)
	}
	return nil
}
