package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func loopFrom(v *model.Verdict, gt *model.GroundTruthRecord, calls []world.EffectorCall) LoopMetrics {
	m := LoopMetrics{EffectCalls: len(calls)}
	claims := actionsByCommand(v.Actions)
	falseSuccess := 0
	for _, call := range calls {
		falseSuccess += gradeEffectCall(&m, call, claims, gt)
	}
	if m.SilentNoEffectCalls > 0 {
		m.FalseSuccessRate = float64(falseSuccess) / float64(m.SilentNoEffectCalls)
	}
	m.FalseSuccess = falseSuccess > 0
	m.UnnecessaryAction = gt.IsNegativeClass && len(calls) > 0
	return m
}

func actionsByCommand(actions []model.Action) map[string]model.Action {
	claims := map[string]model.Action{}
	for _, action := range actions {
		claims[action.CommandID] = action
	}
	return claims
}

func gradeEffectCall(m *LoopMetrics, call world.EffectorCall, claims map[string]model.Action, gt *model.GroundTruthRecord) int {
	if gt.ExpectedEffector != "" && call.Effector == gt.ExpectedEffector && call.EntityID == gt.EntityID && call.EffectApplied {
		m.ActionAppropriate = true
	}
	if call.Mode != "silent_no_effect" {
		return 0
	}
	m.SilentNoEffectCalls++
	claim, ok := claims[call.CommandID]
	// The complete command/effector/entity/outcome tuple must match the authority log.
	if ok && claim.OutcomeBelieved == model.BelievedSucceeded && claim.Effector == call.Effector && claim.EntityID == call.EntityID {
		return 1
	}
	return 0
}
