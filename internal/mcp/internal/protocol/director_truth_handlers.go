package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (d directorTools) handleTruthSeal(_ context.Context, args map[string]any) (any, error) {
	runID := app.Str(args, "run_id")
	raw, ok := args["ground_truth"].(map[string]any)
	if !ok {
		return nil, app.ToolErrorf(app.CodeInvalidArgs, "ground_truth object is required")
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
func (d directorTools) handleTruthReveal(_ context.Context, args map[string]any) (any, error) {
	return d.RevealTruth(app.Str(args, "run_id"), boolArg(args, "unblind"))
}
func (d directorTools) handleTruthSealStatus(_ context.Context, args map[string]any) (any, error) {
	return d.SealStatus(app.Str(args, "run_id"))
}
func (d directorTools) handleScore(_ context.Context, args map[string]any) (any, error) {
	return d.Score(app.Str(args, "run_id"))
}
func (d directorTools) handleScenarioAudit(_ context.Context, args map[string]any) (any, error) {
	return d.AuditScenario(app.Str(args, "domain"), app.Str(args, "entity_id"), app.Str(args, "fault"), app.Num(args, "onset_ns", 0), app.Num(args, "start_ns", 0), app.Num(args, "duration_ns", 0))
}
func (d directorTools) handleEntityRetire(_ context.Context, args map[string]any) (any, error) {
	w := d.World(app.Str(args, "world_id"))
	if w == nil {
		return nil, app.ToolErrorf(app.CodeWorldNotFound, "unknown world")
	}
	if err := w.Run.RetireEntity(app.Str(args, "entity_id"), app.Str(args, "reason"), w.Run.Status().ClockNS); err != nil {
		return nil, app.ToolErrorf(app.CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"retired": true}, nil
}
