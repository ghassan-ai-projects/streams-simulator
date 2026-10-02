package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func loopFrom(v *model.Verdict, gt *model.GroundTruthRecord, calls []world.EffectorCall) LoopMetrics {
	m := LoopMetrics{EffectCalls: len(calls)}
	expected := gt.ExpectedEffector
	// The full tuple must match the call — command id, effector, entity and
	// claimed outcome — the same rule as the online path.
	type claimed struct {
		effector string
		entity   string
		outcome  string
	}
	byCommand := map[string]claimed{}
	for _, a := range v.Actions {
		byCommand[a.CommandID] = claimed{effector: a.Effector, entity: a.EntityID, outcome: a.OutcomeBelieved}
	}
	falseSuccess := 0
	for _, c := range calls {
		if c.Mode == "silent_no_effect" {
			m.SilentNoEffectCalls++
			claim, ok := byCommand[c.CommandID]
			if ok && claim.outcome == model.BelievedSucceeded && claim.effector == c.Effector && claim.entity == c.EntityID {
				falseSuccess++
			}
		}
		if expected != "" && c.Effector == expected && c.EffectApplied {
			m.ActionAppropriate = true
		}
	}
	if m.SilentNoEffectCalls > 0 {
		m.FalseSuccessRate = float64(falseSuccess) / float64(m.SilentNoEffectCalls)
	}
	m.FalseSuccess = falseSuccess > 0
	if gt.IsNegativeClass && len(calls) > 0 {
		m.UnnecessaryAction = true
	}
	return m
}
