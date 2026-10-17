package domain

import (
	"maps"
	"slices"
)

var leaseArguments = []struct{ arg, unit string }{{"lease_ms", "ms"}, {"lease_s", "s"}}

func effectorDocument(k *twinKind, c *commandPlan) map[string]any {
	effector := map[string]any{"name": c.file.Name, "args_schema": c.file.ArgsSchema, "risk_class": c.file.RiskClass,
		"ack": twinAck(), "effect": map[string]any{"state_deltas": stateDeltas(k, c), "dead_time_s": 0, "time_constant_s": 0}, "idempotency_window_s": 60}
	if lease := leaseOf(k, c); lease != nil {
		effector["lease"] = lease
	}
	return effector
}

func twinAck() map[string]any {
	return map[string]any{"failure_modes": []any{map[string]any{"mode": "ok", "probability": 1}}, "latency_ms": map[string]any{"mean": 40, "sigma": 10}}
}

func stateDeltas(k *twinKind, c *commandPlan) []any {
	switch {
	case c.file.SafeStop:
		return safeStopDeltas(k)
	case len(c.choice.Sets) == 0:
		return []any{map[string]any{"state": valueName(k.metrics[0].file.Name), "delta": 0}}
	}
	return assignDeltas(c)
}

func safeStopDeltas(k *twinKind) []any {
	deltas := []any{}
	for _, plan := range k.metrics {
		if eu := plan.file.EURange; k.actuated[plan.file.Name] {
			deltas = append(deltas, map[string]any{"state": valueName(plan.file.Name), "delta": eu.Min - eu.Max})
		}
	}
	return deltas
}

func assignDeltas(c *commandPlan) []any {
	var deltas []any
	for _, arg := range slices.Sorted(maps.Keys(c.choice.Sets)) {
		deltas = append(deltas, map[string]any{"state": valueName(c.choice.Sets[arg]), "assign_from_arg": arg})
	}
	return deltas
}

func leaseOf(k *twinKind, c *commandPlan) map[string]any {
	if len(c.choice.Sets) == 0 {
		return nil
	}
	for _, candidate := range leaseArguments {
		if argumentType(c.file.ArgsSchema, candidate.arg) != "" {
			return map[string]any{"arg": candidate.arg, "unit": candidate.unit, "fallback": "safe_off", "safe_deltas": safeValues(k, c)}
		}
	}
	return nil
}

func safeValues(k *twinKind, c *commandPlan) []any {
	var values []any
	for _, arg := range slices.Sorted(maps.Keys(c.choice.Sets)) {
		metric := c.choice.Sets[arg]
		values = append(values, map[string]any{"state": valueName(metric), "delta": k.plan(metric).file.EURange.Min})
	}
	return values
}
