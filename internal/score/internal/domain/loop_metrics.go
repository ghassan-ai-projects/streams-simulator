package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func loopFrom(v *model.Verdict, gt *model.GroundTruthRecord, calls []world.EffectorCall, matchArg string) LoopMetrics {
	m := LoopMetrics{EffectCalls: len(calls)}
	claims := actionsByCommand(v.Actions)
	falseSuccess := 0
	for _, call := range calls {
		falseSuccess += gradeEffectCall(&m, call, claims, gt, matchArg)
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

func gradeEffectCall(m *LoopMetrics, call world.EffectorCall, claims map[string]model.Action, gt *model.GroundTruthRecord, matchArg string) int {
	if rightCall(call, gt, matchArg) {
		m.ActionAppropriate = true
	}
	if call.Mode != "silent_no_effect" {
		return 0
	}
	m.SilentNoEffectCalls++
	claim, ok := claims[call.CommandID]

	if ok && claim.OutcomeBelieved == model.BelievedSucceeded && claim.Effector == call.Effector && claim.EntityID == call.EntityID {
		return 1
	}
	return 0
}

func rightCall(call world.EffectorCall, gt *model.GroundTruthRecord, matchArg string) bool {
	if gt.ExpectedEffector == "" || call.Effector != gt.ExpectedEffector || call.EntityID != gt.EntityID || !call.EffectApplied {
		return false
	}
	cause, _ := call.Args[matchArg].(string)
	return matchArg == "" || cause == gt.Label
}

func responderArg(ev Evidence, gt *model.GroundTruthRecord) string {
	if ev.Domain == nil {
		return ""
	}
	kind := ev.Domain.InstanceKind(max(ev.Domain.InstanceSetFor(gt.EntityID), 0))
	if kind == nil {
		return ""
	}
	if effector := kind.EffectorDecl(gt.ExpectedEffector); effector != nil && effector.Responder != nil {
		return effector.Responder.MatchArg
	}
	return ""
}
