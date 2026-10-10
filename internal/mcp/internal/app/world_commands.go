package app

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// DescribeWorld reports config, digest, clock and emitted count.
func (d *Director) DescribeWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	status := w.Run.Status()
	return map[string]any{
		"world_id": worldID, "domain": w.Run.Domain().Spec.ID,
		"seed": w.Run.Config.Seed, "clock": model.FormatTime(status.ClockNS),
		"emitted": status.Emitted, "simulated": true,
	}, nil
}

// DestroyWorld finalizes and removes a world.
func (d *Director) DestroyWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := d.finishDestroyedWorld(w, worldID); err != nil {
		return nil, err
	}
	d.mu.Lock()
	delete(d.Worlds, worldID)
	delete(d.byToken, w.Token)
	d.mu.Unlock()
	return map[string]any{"world_id": worldID, "destroyed": true}, nil
}

// Advance moves the clock (sim.clock.advance).
func (d *Director) Advance(ctx context.Context, worldID string, toNS int64, await bool) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	emitted, err := w.Run.Advance(ctx, toNS, await)
	if err != nil {
		return nil, advanceToolError(err)
	}
	status := w.Run.Status()
	return map[string]any{
		"emitted": emitted, "clock": model.FormatTime(status.ClockNS),
		"emitted_total":   status.Emitted,
		"effects_applied": status.ActiveFaults, "simulated": true,
	}, nil
}

// ClockState reports the clock and queue.
func (d *Director) ClockState(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	status := w.Run.Status()
	return map[string]any{
		"clock":             model.FormatTime(status.ClockNS),
		"next_scheduled_ns": status.NextEventNS,
		"pending_effects":   status.PendingEffects,
	}, nil
}

// InjectFault applies a world fault.
func (d *Director) InjectFault(worldID, entityID, fault string, onsetNS int64, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	fid, err := w.Run.InjectFault(entityID, fault, onsetNS, params)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"fault_id": fid, "simulated": true}, nil
}

// ClearFault clears a fault.
func (d *Director) ClearFault(worldID, faultID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearFault(faultID, 0); err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

// ListFaults is director-only: active faults, with ids and onsets.
func (d *Director) ListFaults(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	return map[string]any{"faults": w.Run.Faults()}, nil
}

// ApplyPerturb activates a delivery perturbation.
func (d *Director) ApplyPerturb(worldID, name string, params map[string]any, fromNS, untilNS int64) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.ApplyPerturb(name, params, fromNS, untilNS)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"perturb_id": id}, nil
}

// ClearPerturb deactivates a perturbation.
func (d *Director) ClearPerturb(worldID, id string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearPerturb(id); err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

// EnvInject records an environment fault against a configured target.
func (d *Director) EnvInject(worldID, target, fault string, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.EnvInject(target, fault, params, w.Run.Status().ClockNS)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"env_id": id}, nil
}

func (d *Director) finishDestroyedWorld(w *WorldRecord, worldID string) error {
	if _, ended := d.runLifecycle(w); !ended {
		dir := filepath.Join(d.OutDir, worldID)
		_, err := w.Run.End(dir)
		// An End that failed still finished the run; leaving it open would
		// make every retry fail with "already finished".
		d.markRunEnded(w)
		if err != nil {
			return errTool(CodeDomainInvalid, "%v", err)
		}
	}
	return nil
}

func advanceToolError(err error) error {
	switch {
	case errors.Is(err, run.ErrConsumerNotQuiesced):
		return errTool(CodeConsumerNotQuiesced, "%v", err)
	case errors.Is(err, world.ErrClockBackwards):
		return errTool(CodeClockBackwards, "%v", err)
	}
	return errTool(CodeDomainInvalid, "%v", err)
}
