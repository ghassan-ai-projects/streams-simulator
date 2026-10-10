package domain

import (
	"encoding/json"
)

func (w *World) recordCall(inv invocation, outcome callOutcome) {
	call := &EffectorCall{CommandID: inv.commandID, Effector: inv.effector, EntityID: inv.entityID, WorldID: w.ID, Args: inv.args, AtNS: inv.atNS}
	call.recordOutcome(outcome)
	call.ResultDigest = resultDigest(call)
	w.effectorCalls = append(w.effectorCalls, *call)
}

func resultDigest(c *EffectorCall) string {
	// Best-effort short digest of the outcome; digests for the run artifact
	// are computed by the run layer from canonical JSON.
	b, _ := json.Marshal(map[string]any{
		"mode": c.Mode, "accepted": c.Accepted, "reason": c.Reason,
		"effect_applied": c.EffectApplied,
	})
	return string(b)
}

// EffectorCalls returns the effector call log (the authority when scoring
// actions).
func (w *World) EffectorCalls() []EffectorCall {
	out := make([]EffectorCall, len(w.effectorCalls))
	copy(out, w.effectorCalls)
	return out
}

func (call *EffectorCall) recordOutcome(outcome callOutcome) {
	call.Mode = outcome.mode
	call.Accepted = outcome.accepted
	call.InterlockRefused = outcome.interlock
	call.Reason = outcome.reason
	call.AckLatencyMS = outcome.latencyMS
	call.EffectApplied = outcome.effectApplied
}
