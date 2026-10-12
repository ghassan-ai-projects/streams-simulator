package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestScheduleRegimeAndPeakRulesAreLocated(t *testing.T) {
	t.Parallel()
	cases := []violationCase{
		{"schedule points out of order", func(d map[string]any) {
			put(d, 80000.0, "kinds", 0, "dynamics", 0, "f0", "schedule", "points", 1, "at_s")
		}, "/kinds/0/dynamics/0/f0/schedule/points/2/at_s"},
		{"a schedule point beyond its period", func(d map[string]any) {
			put(d, 90000.0, "kinds", 0, "dynamics", 0, "f0", "schedule", "points", 2, "at_s")
		}, "/kinds/0/dynamics/0/f0/schedule/points/2/at_s"},
		{"a peak over two inputs", func(d map[string]any) { put(d, []any{"load", "tariff"}, "kinds", 0, "dynamics", 2, "f1", "inputs") }, "/kinds/0/dynamics/2/f1/inputs"},
		{"a peak block that is not whole steps", func(d map[string]any) { put(d, 930.0, "kinds", 0, "dynamics", 2, "f1", "block_s") }, "/kinds/0/dynamics/2/f1/block_s"},
		{"a peak period that is not whole blocks", func(d map[string]any) { put(d, 1000.0, "kinds", 0, "dynamics", 2, "f1", "period_s") }, "/kinds/0/dynamics/2/f1/period_s"},
		{"a regime selector that is no state", func(d map[string]any) { put(d, "ambient", "kinds", 0, "dynamics", 3, "f1", "selector") }, "/kinds/0/dynamics/3/f1/selector"},
		{"a regime without its above lag", func(d map[string]any) { delete(at(d, "kinds", 0, "dynamics", 3, "f1").(map[string]any), "above") }, "/kinds/0/dynamics/3/f1/above"},
	}
	runFixtureViolations(t, testsupport.UtilityDomain(), cases)
}

func TestTripAntiRecycleBoundaryAndPreconditionRulesAreLocated(t *testing.T) {
	t.Parallel()
	runFixtureViolations(t, testsupport.UtilityDomain(), []violationCase{
		{"a trip on an undeclared state", func(d map[string]any) { put(d, "pressure", "kinds", 1, "dynamics", 2, "f1", "when", "state") }, "/kinds/1/dynamics/2/f1/when/state"},
		{"a comparator outside the closed set", func(d map[string]any) { put(d, "approx", "kinds", 1, "dynamics", 2, "f1", "when", "operator") }, "/kinds/1/dynamics/2/f1/when/operator"},
		{"a latched trip without a reset", func(d map[string]any) { delete(at(d, "kinds", 1, "dynamics", 2, "f1").(map[string]any), "reset_on") }, "/kinds/1/dynamics/2/f1/reset_on"},
		{"an anti-recycle on an undeclared state", func(d map[string]any) { put(d, "motor", "kinds", 1, "dynamics", 4, "f1", "watch") }, "/kinds/1/dynamics/4/f1/watch"},
		{"a boundary on a state that is no trip", func(d map[string]any) { put(d, "flow", "kinds", 1, "boundaries", 0, "state") }, "/kinds/1/boundaries/0/state"},
		{"a boundary declared twice", func(d map[string]any) { put(d, "low_flow_trip", "kinds", 1, "boundaries", 1, "name") }, "/kinds/1/boundaries/1/name"},
		{"a tripped value equal to the initial value", func(d map[string]any) { put(d, 0.0, "kinds", 1, "dynamics", 3, "f1", "tripped_value") }, "/kinds/1/dynamics/3/f1/tripped_value"},
		{"a severity outside the classes", func(d map[string]any) { put(d, "fatal", "kinds", 1, "boundaries", 0, "severity") }, "/kinds/1/boundaries/0/severity"},
		{"a precondition on an undeclared state", func(d map[string]any) { put(d, "door", "kinds", 1, "effectors", 1, "requires", 0, "state") }, "/kinds/1/effectors/1/requires/0/state"},
	})
}
