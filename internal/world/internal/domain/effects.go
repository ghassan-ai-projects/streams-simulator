package domain

import (
	"encoding/json"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// applyAutonomousAction applies the interlock's autonomous state delta.
func (w *World) applyAutonomousAction(entityID, state string, delta float64, atNS int64) {
	k := &kick{entity: entityID, state: state, delta: delta, startNS: atNS, tauNS: 0}
	key := driverKey{entity: entityID, state: state}
	w.kicks[key] = append(w.kicks[key], k)
	w.schedule(kindEffectStart, entityID, state, atNS, k)
}

// applyEffect schedules the effect's state deltas as kicks, in the real
// world or in the shadow world (silent_no_effect).
func (w *World) applyEffect(entityID string, eff *model.Effector, args map[string]any, atNS int64, scale float64, shadow bool) {
	start := atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	tau := eff.Effect.TimeConstantS * secondsPerNS
	for _, delta := range eff.Effect.StateDeltas {
		contribution, assign, ok := effectContribution(delta, args)
		if !ok {
			continue
		}
		contribution *= scale
		kick := &kick{entity: entityID, state: delta.State, delta: contribution, assign: assign, startNS: start, tauNS: tau}
		w.scheduleEffectKick(kick, shadow)
	}
}

func argFloat(args map[string]any, name string) (float64, bool) {
	value, ok := args[name]
	if !ok {
		return 0, false
	}
	return argumentFloat(value)
}

func argumentFloat(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	case int64:
		return float64(number), true
	case int:
		return float64(number), true
	}
	return 0, false
}

func (w *World) scheduleEffectKick(kick *kick, shadow bool) {
	key := driverKey{entity: kick.entity, state: kick.state}
	if shadow {
		w.shadow[key] = append(w.shadow[key], kick)
	} else {
		w.kicks[key] = append(w.kicks[key], kick)
	}
	w.schedule(kindEffectStart, kick.entity, kick.state, kick.startNS, kick)
}

func effectContribution(delta model.StateDelta, args map[string]any) (float64, bool, bool) {
	value := additiveEffectValue(delta, args)
	if delta.AssignFromArg != "" {
		argument, ok := argFloat(args, delta.AssignFromArg)
		if !ok {
			return 0, false, false
		}
		return argument, true, true
	}
	return value, false, true
}

func additiveEffectValue(delta model.StateDelta, args map[string]any) float64 {
	value := delta.Delta
	if delta.FromArg != "" {
		if argument, ok := argFloat(args, delta.FromArg); ok {
			value = argument
		}
	}
	return value
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
