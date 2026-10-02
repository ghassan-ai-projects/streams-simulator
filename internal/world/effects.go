package world

import (
	"encoding/json"

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
	for _, d := range eff.Effect.StateDeltas {
		delta := d.Delta
		assign := false
		if d.FromArg != "" {
			if v, ok := argFloat(args, d.FromArg); ok {
				delta = v
			}
		}
		if d.AssignFromArg != "" {
			v, ok := argFloat(args, d.AssignFromArg)
			if !ok {
				continue
			}
			delta = v
			assign = true
		}
		delta *= scale
		k := &kick{entity: entityID, state: d.State, delta: delta, assign: assign, startNS: start, tauNS: tau}
		key := driverKey{entity: entityID, state: d.State}
		if shadow {
			w.shadow[key] = append(w.shadow[key], k)
		} else {
			w.kicks[key] = append(w.kicks[key], k)
		}
		w.schedule(kindEffectStart, entityID, d.State, start, k)
	}
}

func argFloat(args map[string]any, name string) (float64, bool) {
	if args == nil {
		return 0, false
	}
	v, ok := args[name]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}
