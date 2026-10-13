package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pidDynamics(pid map[string]any, clamp map[string]any) map[string]any {
	return map[string]any{"target": "avg_flow", "tier": "F1", "dt_ms": 1000, "f1": map[string]any{"form": "pid", "pid": pid, "clamp": clamp}}
}

func TestAPIDControllerIsValidatedByItsReadsGainsAndLimits(t *testing.T) {
	t.Parallel()
	limits := map[string]any{"min": 0.0, "max": 100.0}
	pid := func(extra map[string]any) map[string]any {
		out := map[string]any{"pv": "level", "sp": "draw", "kp": 1.0, "action": "reverse"}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	cases := []struct {
		name     string
		dynamics map[string]any
		pointer  string
	}{
		{"a valid controller", pidDynamics(pid(nil), limits), ""},
		{"a manual mode", pidDynamics(pid(map[string]any{"manual": "inflow", "manual_output": "draw"}), limits), ""},
		{"an undeclared process variable", pidDynamics(pid(map[string]any{"pv": "temp"}), limits), "/kinds/1/dynamics/3/f1/pid/pv"},
		{"an undeclared setpoint", pidDynamics(pid(map[string]any{"sp": "temp"}), limits), "/kinds/1/dynamics/3/f1/pid/sp"},
		{"manual without an output", pidDynamics(pid(map[string]any{"manual": "inflow"}), limits), "/kinds/1/dynamics/3/f1/pid/manual"},
		{"no gain", pidDynamics(pid(map[string]any{"kp": 0.0}), limits), "/kinds/1/dynamics/3/f1/pid/kp"},
		{"an open output", pidDynamics(pid(nil), map[string]any{"min": 0.0}), "/kinds/1/dynamics/3/f1/clamp"},
		{"an empty output range", pidDynamics(pid(nil), map[string]any{"min": 5.0, "max": 5.0}), "/kinds/1/dynamics/3/f1/clamp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			violationIn(t, testsupport.CoupledDomain(), func(d map[string]any) { put(d, tc.dynamics, "kinds", 1, "dynamics", 3) }, tc.pointer)
		})
	}
}
