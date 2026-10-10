package domain

import (
	"encoding/json"
)

func (w *World) recordCall(inv invocation, outcome callOutcome) {
	call := &EffectorCall{CommandID: inv.commandID, Effector: inv.effector, EntityID: inv.entityID, WorldID: w.ID, Args: cloneArgs(inv.args), AtNS: inv.atNS}
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
	for i := range out {
		out[i].Args = cloneArgs(out[i].Args)
	}
	return out
}

// cloneArgs deep-copies effector arguments (objects and arrays), so neither
// the caller that issued a command nor a reader of the call log can change
// what the log says was asked.
func cloneArgs(args map[string]any) map[string]any {
	if args == nil {
		return nil
	}
	out := make(map[string]any, len(args))
	// determinism-safe: copies a map into a map.
	for key, value := range args {
		out[key] = cloneArgValue(value)
	}
	return out
}

func cloneArgValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneArgs(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneArgValue(item)
		}
		return out
	}
	return value
}

func (call *EffectorCall) recordOutcome(outcome callOutcome) {
	call.Mode = outcome.mode
	call.Accepted = outcome.accepted
	call.InterlockRefused = outcome.interlock
	call.Reason = outcome.reason
	call.AckLatencyMS = outcome.latencyMS
	call.EffectApplied = outcome.effectApplied
}
