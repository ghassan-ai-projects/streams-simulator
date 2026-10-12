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
