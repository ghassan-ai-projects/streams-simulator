package app

import (
	"context"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (d *Director) handleTruthSeal(_ context.Context, args map[string]any) (any, error) {
	runID := str(args, "run_id")
	raw, ok := args["ground_truth"].(map[string]any)
	if !ok {
		return nil, errTool(CodeInvalidArgs, "ground_truth object is required")
	}
	rec := &model.GroundTruthRecord{}
	if err := decodeToolRecord(raw, rec); err != nil {
		return nil, err
	}
	if err := d.SealTruth(runID, rec); err != nil {
		return nil, err
	}
	return map[string]any{"run_id": runID, "sealed": true}, nil
}
func (d *Director) handleTruthReveal(_ context.Context, args map[string]any) (any, error) {
	return d.RevealTruth(str(args, "run_id"), boolArg(args, "unblind"))
}
func (d *Director) handleTruthSealStatus(_ context.Context, args map[string]any) (any, error) {
	return d.SealStatus(str(args, "run_id"))
}
func (d *Director) handleScore(_ context.Context, args map[string]any) (any, error) {
	return d.Score(str(args, "run_id"))
}
func (d *Director) handleScenarioAudit(_ context.Context, args map[string]any) (any, error) {
	return d.AuditScenario(str(args, "domain"), str(args, "entity_id"), str(args, "fault"), num(args, "onset_ns", 0), num(args, "start_ns", 0), num(args, "duration_ns", 0))
}
func (d *Director) handleEntityRetire(_ context.Context, args map[string]any) (any, error) {
	w := d.World(str(args, "world_id"))
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world")
	}
	if err := w.Run.RetireEntity(str(args, "entity_id"), str(args, "reason"), w.Run.World.Clock()); err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"retired": true}, nil
}
