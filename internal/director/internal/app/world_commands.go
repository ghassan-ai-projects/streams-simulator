package app

import (
	"context"
	"errors"
	"path/filepath"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (d *Director) DescribeWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	status := w.Run.Status()
	return map[string]any{
		"world_id": worldID, "domain": w.Run.Domain().Spec.ID,
		"seed": w.Run.Config.Seed, "clock": model.FormatTime(status.ClockNS),
		"emitted": status.Emitted, "simulated": true,
	}, nil
}

func (d *Director) DestroyWorld(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
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

func (d *Director) Advance(ctx context.Context, worldID string, toNS int64, await bool) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	emitted, err := w.Run.Advance(ctx, toNS, await)
	if err != nil {
		return nil, advanceToolError(err)
	}
	return advanceResult(emitted, w.Run.Status()), nil
}

func (d *Director) AdvanceBy(ctx context.Context, worldID string, byNS int64, await bool) (map[string]any, error) {
	toNS, err := d.relativeClockTarget(worldID, byNS)
	if err != nil {
		return nil, err
	}
	return d.Advance(ctx, worldID, toNS, await)
}

func (d *Director) relativeClockTarget(worldID string, byNS int64) (int64, error) {
	if byNS < 0 {
		return 0, rules.ToolErrorf(rules.CodeInvalidArgs, "by_ns must be non-negative")
	}
	w := d.World(worldID)
	if w == nil {
		return 0, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	const maxInt64 = int64(1<<63 - 1)
	if byNS > maxInt64-w.Run.Status().ClockNS {
		return 0, rules.ToolErrorf(rules.CodeInvalidArgs, "by_ns overflows the world clock")
	}
	return w.Run.Status().ClockNS + byNS, nil
}

func (d *Director) RetireEntity(worldID, entityID, reason string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world")
	}
	if err := w.Run.RetireEntity(entityID, reason, w.Run.Status().ClockNS); err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"retired": true}, nil
}

func advanceResult(emitted int, status run.WorldStatus) map[string]any {
	return map[string]any{
		"emitted": emitted, "clock": model.FormatTime(status.ClockNS),
		"emitted_total":   status.Emitted,
		"effects_applied": status.ActiveFaults, "simulated": true,
	}
}

func (d *Director) ClockState(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	status := w.Run.Status()
	return map[string]any{
		"clock":             model.FormatTime(status.ClockNS),
		"next_scheduled_ns": status.NextEventNS,
		"pending_effects":   status.PendingEffects,
	}, nil
}

func (d *Director) InjectFault(worldID, entityID, fault string, onsetNS int64, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	fid, err := w.Run.InjectFault(entityID, fault, onsetNS, params)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"fault_id": fid, "simulated": true}, nil
}

func (d *Director) ClearFault(worldID, faultID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearFault(faultID, 0); err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

func (d *Director) ListFaults(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	return map[string]any{"faults": w.Run.Faults()}, nil
}

func (d *Director) ApplyPerturb(worldID, name string, params map[string]any, fromNS, untilNS int64) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.ApplyPerturb(name, params, fromNS, untilNS)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"perturb_id": id}, nil
}

func (d *Director) ClearPerturb(worldID, id string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	if err := w.Run.ClearPerturb(id); err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"cleared": true}, nil
}

func (d *Director) EnvInject(worldID, target, fault string, params map[string]any) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	id, err := w.Run.EnvInject(target, fault, params, w.Run.Status().ClockNS)
	if err != nil {
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"env_id": id}, nil
}

func (d *Director) finishDestroyedWorld(w *WorldRecord, worldID string) error {
	if _, ended := d.runLifecycle(w); !ended {
		dir := filepath.Join(d.OutDir, worldID)
		_, err := w.Run.End(dir)

		d.markRunEnded(w)
		if err != nil {
			return rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
		}
	}
	return nil
}

func advanceToolError(err error) error {
	switch {
	case errors.Is(err, run.ErrConsumerNotQuiesced):
		return rules.ToolErrorf(rules.CodeConsumerNotQuiesced, "%v", err)
	case errors.Is(err, world.ErrClockBackwards):
		return rules.ToolErrorf(rules.CodeClockBackwards, "%v", err)
	}
	return rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
}
