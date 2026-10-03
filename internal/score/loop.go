package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func loop(r *run.Run, gt *model.GroundTruthRecord) LoopMetrics {
	v := r.Verdict()
	calls := r.World.EffectorCalls()
	m := loopFrom(v, gt, calls)
	if gt.ExpectedEffector != "" && len(calls) > 0 {
		m.Resolved, m.TimeToResolutionNS = resolveTime(r, gt)
	}
	if gt.DeadlineNS > 0 {
		m.DeadlineAdhered = actsBeforeDeadline(calls, gt) && !m.FalseSuccess && !m.UnnecessaryAction
	}
	return m
}

// resolveTime scans the world after the first successful expected-effector
// call for the primary affected state to return toward its pre-fault level.
func resolveTime(r *run.Run, gt *model.GroundTruthRecord) (bool, int64) {
	state, faulted, deviation, ok := recoveryTarget(r, gt)
	if !ok {
		return false, 0
	}
	start := firstSuccessfulCall(r.World.EffectorCalls(), gt.ExpectedEffector)
	if start == 0 {
		return false, 0
	}
	return recoveryAfter(r, gt.EntityID, state, start, faulted+0.5*deviation, deviation)
}

func actsBeforeDeadline(calls []world.EffectorCall, gt *model.GroundTruthRecord) bool {
	for _, call := range calls {
		if call.Effector == gt.ExpectedEffector && call.EntityID == gt.EntityID && call.EffectApplied && call.AtNS <= gt.DeadlineNS {
			return true
		}
	}
	return false
}

func recoveryTarget(r *run.Run, gt *model.GroundTruthRecord) (string, float64, float64, bool) {
	fault := faultFor(r, gt.Label)
	if fault == nil || len(fault.Affects) == 0 {
		return "", 0, 0, false
	}
	state := fault.Affects[0].State
	faulted, deviation, ok := recoveryLevels(r, gt, state)
	return state, faulted, deviation, ok
}

func recoveryLevels(r *run.Run, gt *model.GroundTruthRecord, state string) (float64, float64, bool) {
	// Onset samples may already carry the fault; baseline must precede onset.
	baseline, ok := historyValueBefore(r, gt.EntityID, state, gt.InjectionTimeNS)
	if !ok {
		return 0, 0, false
	}
	faulted, ok := historyValue(r, gt.EntityID, state, gt.FirstObservableTimeNS+60*1e9)
	if !ok {
		return 0, 0, false
	}
	deviation := baseline - faulted
	return faulted, deviation, deviation != 0
}

func firstSuccessfulCall(calls []world.EffectorCall, effector string) int64 {
	start := int64(0)
	for _, call := range calls {
		if call.Effector == effector && call.EffectApplied && (start == 0 || call.AtNS < start) {
			start = call.AtNS
		}
	}
	return start
}

func recoveryAfter(r *run.Run, entity, state string, start int64, level, deviation float64) (bool, int64) {
	// Recovery means returning within 50% of the original deviation.
	for _, snap := range r.History() {
		if snap.Entity != entity || snap.TimeNS < start {
			continue
		}
		value := snap.States[state]
		if (deviation > 0 && value >= level) || (deviation < 0 && value <= level) {
			return true, snap.TimeNS - start
		}
	}
	return false, 0
}
