package world

import (
	"encoding/json"
	"sort"
)

func (w *World) recordCall(effector, entityID, commandID string, args map[string]any, atNS int64, mode string, accepted, interlock bool, reason string, latency float64, effectApplied bool) {
	call := &EffectorCall{
		CommandID:        commandID,
		Effector:         effector,
		EntityID:         entityID,
		WorldID:          w.ID,
		Args:             args,
		AtNS:             atNS,
		Mode:             mode,
		Accepted:         accepted,
		InterlockRefused: interlock,
		Reason:           reason,
		AckLatencyMS:     latency,
		EffectApplied:    effectApplied,
	}
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

// HiddenStateSnapshot returns the true hidden state values of every live
// entity at time t (director-only; feeds world_state_history).
func (w *World) HiddenStateSnapshot(t int64) map[string]map[string]float64 {
	out := map[string]map[string]float64{}
	for _, id := range w.entityOrder {
		ent := w.entities[id]
		if ent == nil || !ent.alive {
			continue
		}
		row := map[string]float64{}
		for _, name := range w.sortedStateNames() {
			row[name] = w.stateAt(id, name, t)
		}
		out[id] = row
	}
	return out
}

// sortedStateNames returns the declared state names in a stable order.
func (w *World) sortedStateNames() []string {
	names := w.Spec.StateNames()
	sort.Strings(names)
	return names
}

// PendingKicks is the number of scheduled-but-unapplied effect kicks.
func (w *World) PendingKicks() int {
	n := 0
	for _, key := range w.sortedKickStates() {
		for _, k := range w.kicks[key] {
			if !k.applied {
				n++
			}
		}
	}
	return n
}

// sortedKickStates returns the states with pending kicks in a stable order.
func (w *World) sortedKickStates() []driverKey {
	var out []driverKey
	// determinism-safe: collected here, sorted below before any output.
	for key := range w.kicks {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].entity != out[j].entity {
			return out[i].entity < out[j].entity
		}
		return out[i].state < out[j].state
	})
	return out
}
