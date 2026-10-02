package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func loop(r *run.Run, gt *model.GroundTruthRecord) LoopMetrics {
	v := r.Verdict()
	calls := r.World.EffectorCalls()
	m := loopFrom(v, gt, calls)
	// Resolution: did the primary affected state recover after the action?
	if gt.ExpectedEffector != "" && len(calls) > 0 {
		res, t := resolveTime(r, gt)
		m.Resolved = res
		m.TimeToResolutionNS = t
	}
	if gt.DeadlineNS > 0 {
		m.DeadlineAdhered = false
		for _, c := range calls {
			if c.Effector == gt.ExpectedEffector && c.EntityID == gt.EntityID && c.EffectApplied && c.AtNS <= gt.DeadlineNS {
				m.DeadlineAdhered = true
				break
			}
		}
		m.DeadlineAdhered = m.DeadlineAdhered && !m.FalseSuccess && !m.UnnecessaryAction
	}
	return m
}

// resolveTime scans the world after the first successful expected-effector
// call for the primary affected state to return toward its pre-fault level.
func resolveTime(r *run.Run, gt *model.GroundTruthRecord) (bool, int64) {
	fault := faultFor(r, gt.Label)
	if fault == nil || len(fault.Affects) == 0 {
		return false, 0
	}
	state := fault.Affects[0].State
	// Pre-fault baseline: the state just before onset. The sample at or
	// after the onset may already carry the fault when the onset lands on
	// an emission boundary, so the baseline is the latest sample strictly
	// before it.
	base, ok := historyValueBefore(r, gt.EntityID, state, gt.InjectionTimeNS)
	if !ok {
		return false, 0
	}
	faulted, ok := historyValue(r, gt.EntityID, state, gt.FirstObservableTimeNS+60*1e9)
	if !ok {
		return false, 0
	}
	deviation := base - faulted
	if deviation == 0 {
		return false, 0
	}
	// First successful expected-effector call.
	start := int64(0)
	for _, c := range r.World.EffectorCalls() {
		if c.Effector == gt.ExpectedEffector && c.EffectApplied && (start == 0 || c.AtNS < start) {
			start = c.AtNS
		}
	}
	if start == 0 {
		return false, 0
	}
	// Recovery: the state returned within 50% of the deviation.
	recoverLevel := faulted + 0.5*deviation
	for _, snap := range r.History() {
		if snap.Entity != gt.EntityID || snap.TimeNS < start {
			continue
		}
		v := snap.States[state]
		if (deviation > 0 && v >= recoverLevel) || (deviation < 0 && v <= recoverLevel) {
			return true, snap.TimeNS - start
		}
	}
	return false, 0
}
