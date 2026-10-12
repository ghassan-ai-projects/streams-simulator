package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (d directorTools) handleTruthSeal(_ context.Context, args map[string]any) (any, error) {
	runID := stringArg(args, "run_id")
	raw, ok := args["ground_truth"].(map[string]any)
	if !ok {
		return nil, director.ToolErrorf(director.CodeInvalidArgs, "ground_truth object is required")
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
	return d.RevealTruth(stringArg(args, "run_id"), boolArg(args, "unblind"))
}
func (d directorTools) handleTruthSealStatus(_ context.Context, args map[string]any) (any, error) {
	return d.SealStatus(stringArg(args, "run_id"))
}
func (d directorTools) handleScore(_ context.Context, args map[string]any) (any, error) {
	return d.Score(stringArg(args, "run_id"))
}
func (d directorTools) handleScenarioAudit(_ context.Context, args map[string]any) (any, error) {
	return d.AuditScenario(stringArg(args, "domain"), stringArg(args, "entity_id"), stringArg(args, "fault"), numberArg(args, "onset_ns", 0), numberArg(args, "start_ns", 0), numberArg(args, "duration_ns", 0))
}
func (d directorTools) handleEntityRetire(_ context.Context, args map[string]any) (any, error) {
	return d.RetireEntity(stringArg(args, "world_id"), stringArg(args, "entity_id"), stringArg(args, "reason"))
}
