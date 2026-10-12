package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) followUpCall(eff *model.Effector, inv invocation) {
	if eff.Responder != nil {
		w.dispatchResponder(eff.Responder, inv)
	}
	if eff.Lease != nil {
		w.armLease(eff, inv)
	}
}

func (w *World) dispatchResponder(r *model.Responder, inv invocation) {
	arrival := inv.atNS + int64(r.DelayS*secondsPerNS)
	cause, _ := inv.args[r.MatchArg].(string)
	matched := w.activeInjections(inv.entityID, cause, arrival)
	deltas := r.OnMiss
	if len(matched) > 0 {
		deltas = r.OnMatch
	}
	for _, id := range matched {
		if r.ClearFault {
			w.faultsByID[id].clearedNS = arrival
		}
	}
	w.scheduleDeltas(inv.entityID, deltas, inv.args, arrival, 1)
}

func (w *World) activeInjections(entityID, faultID string, at int64) []string {
	var ids []string
	for _, id := range slices.Sorted(maps.Keys(w.faultsByID)) {
		active := w.faultsByID[id]
		if active.entity == entityID && active.fault.ID == faultID && active.activeAt(at) {
			ids = append(ids, id)
		}
	}
	return ids
}

func (w *World) scheduleDeltas(entityID string, deltas []model.StateDelta, args map[string]any, at int64, sign float64) []*kick {
	var kicks []*kick
	for _, delta := range deltas {
		contribution, assign, ok := effectContribution(delta, args)
		if !ok {
			continue
		}
		k := &kick{entity: entityID, state: delta.State, delta: sign * contribution, assign: assign, startNS: at}
		w.scheduleEffectKick(k, false)
		kicks = append(kicks, k)
	}
	return kicks
}

func (w *World) armLease(eff *model.Effector, inv invocation) {
	seconds, ok := argFloat(inv.args, eff.Lease.Arg)
	if !ok {
		return
	}
	key := driverKey{entity: inv.entityID, state: eff.Name}
	carried := w.withdrawLease(key, inv.atNS)
	lapse := inv.atNS + int64((eff.Effect.DeadTimeS+seconds)*secondsPerNS)
	switch eff.Lease.Fallback {
	case model.FallbackSafeOff:
		w.leases[key] = w.scheduleSafeState(inv.entityID, eff.Lease.SafeDeltas, lapse)
	case model.FallbackRevert:
		w.leases[key] = w.scheduleRevert(eff, inv, lapse, carried)
	}
}

func (w *World) scheduleRevert(eff *model.Effector, inv invocation, lapse int64, carried map[string]float64) []*kick {
	totals := maps.Clone(carried)
	for _, delta := range eff.Effect.StateDeltas {
		if contribution, _, ok := effectContribution(delta, inv.args); ok {
			totals[delta.State] += contribution
		}
	}
	kicks := make([]*kick, 0, len(totals))
	for _, state := range slices.Sorted(maps.Keys(totals)) {
		k := &kick{entity: inv.entityID, state: state, delta: -totals[state], startNS: lapse, tauNS: eff.Effect.TimeConstantS * secondsPerNS}
		w.scheduleEffectKick(k, false)
		kicks = append(kicks, k)
	}
	return kicks
}

func (w *World) scheduleSafeState(entityID string, deltas []model.StateDelta, at int64) []*kick {
	kicks := make([]*kick, 0, len(deltas))
	for _, delta := range deltas {
		k := &kick{entity: entityID, state: delta.State, delta: delta.Delta, assign: true, startNS: at}
		w.scheduleEffectKick(k, false)
		kicks = append(kicks, k)
	}
	return kicks
}

func (w *World) withdrawLease(key driverKey, at int64) map[string]float64 {
	carried := map[string]float64{}
	for _, pending := range w.leases[key] {
		if pending.startNS > at {
			w.withdrawKick(pending)
			carried[pending.state] -= pending.delta
		}
	}
	delete(w.leases, key)
	return carried
}

func (w *World) withdrawKick(k *kick) {
	key := driverKey{entity: k.entity, state: k.state}
	w.kicks[key] = slices.DeleteFunc(w.kicks[key], func(candidate *kick) bool { return candidate == k })
	k.applied = true
}
