package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func curveDynamics(table map[string]any, inputs ...any) map[string]any {
	return map[string]any{"target": "avg_flow", "tier": "F1", "f1": map[string]any{"form": "curve", "inputs": inputs, "table": table}}
}

func TestCurveTablesAndProductsAreValidatedByShape(t *testing.T) {
	t.Parallel()
	oneInput := map[string]any{"x": []any{0.0, 1.0}, "values": []any{0.0, 2.0}}
	cases := []struct {
		name     string
		dynamics map[string]any
		pointer  string
	}{
		{"a one-input curve with no step", curveDynamics(oneInput, "inflow"), ""},
		{"a two-input curve", curveDynamics(map[string]any{"x": []any{0.0, 1.0}, "y": []any{0.0, 5.0, 9.0}, "grid": []any{[]any{1.0, 2.0, 3.0}, []any{4.0, 5.0, 6.0}}}, "inflow", "draw"), ""},
		{"an x axis that does not increase", curveDynamics(map[string]any{"x": []any{1.0, 1.0}, "values": []any{0.0, 2.0}}, "inflow"), "/kinds/1/dynamics/3/f1/table/x"},
		{"a value per x missing", curveDynamics(map[string]any{"x": []any{0.0, 1.0, 2.0}, "values": []any{0.0, 2.0}}, "inflow"), "/kinds/1/dynamics/3/f1/table/values"},
		{"a short grid row", curveDynamics(map[string]any{"x": []any{0.0, 1.0}, "y": []any{0.0, 5.0}, "grid": []any{[]any{1.0, 2.0}, []any{4.0, 5.0, 6.0}}}, "inflow", "draw"), "/kinds/1/dynamics/3/f1/table/grid"},
		{"three inputs", curveDynamics(oneInput, "inflow", "draw", "level"), "/kinds/1/dynamics/3/f1/inputs"},
		{"a product of one input", map[string]any{"target": "avg_flow", "tier": "F1", "f1": map[string]any{"form": "product", "inputs": []any{"inflow"}}}, "/kinds/1/dynamics/3/f1/inputs"},
		{"a product of two inputs", map[string]any{"target": "avg_flow", "tier": "F1", "f1": map[string]any{"form": "product", "inputs": []any{"inflow", "draw"}}}, ""},
		{"an algebraic loop", map[string]any{"target": "avg_flow", "tier": "F1", "f1": map[string]any{"form": "sum", "inputs": []any{"avg_flow"}}}, "/kinds/1/dynamics/3/target"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			violationIn(t, testsupport.CoupledDomain(), func(d map[string]any) { put(d, tc.dynamics, "kinds", 1, "dynamics", 3) }, tc.pointer)
		})
	}
}
