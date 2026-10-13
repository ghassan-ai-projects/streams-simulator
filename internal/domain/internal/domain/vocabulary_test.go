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
		{"a utc offset without a period", func(d map[string]any) {
			schedule := at(d, "kinds", 0, "dynamics", 0, "f0", "schedule").(map[string]any)
			delete(schedule, "period_s")
			schedule["utc_offset_s"] = 3600.0
		}, "/kinds/0/dynamics/0/f0/schedule/utc_offset_s"},
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

func TestSequencerResponderAndLeaseRulesAreLocated(t *testing.T) {
	t.Parallel()
	runFixtureViolations(t, testsupport.UtilityDomain(), []violationCase{
		{"a band on an undeclared state", func(d map[string]any) {
			put(d, "band_mid", "kinds", 2, "dynamics", 2, "f1", "stage_up", "threshold_state")
		}, "/kinds/2/dynamics/2/f1/stage_up/threshold_state"},
		{"a sequencer without stages", func(d map[string]any) { delete(at(d, "kinds", 2, "dynamics", 2, "f1").(map[string]any), "stages") }, "/kinds/2/dynamics/2/f1/stages"},
		{"a responder matching an undeclared argument", func(d map[string]any) { put(d, "reason", "kinds", 2, "effectors", 2, "responder", "match_arg") }, "/kinds/2/effectors/2/responder/match_arg"},
		{"a responder acting on an undeclared state", func(d map[string]any) {
			put(d, "valve", "kinds", 2, "effectors", 2, "responder", "on_match", 0, "state")
		}, "/kinds/2/effectors/2/responder/on_match/0/state"},
		{"a lease reading a string argument", func(d map[string]any) {
			put(d, map[string]any{"type": "string"}, "kinds", 2, "effectors", 1, "args_schema", "properties", "lease_s")
		}, "/kinds/2/effectors/1/lease/arg"},
		{"a safe-off lease without a safe state", func(d map[string]any) {
			delete(at(d, "kinds", 2, "effectors", 1, "lease").(map[string]any), "safe_deltas")
		}, "/kinds/2/effectors/1/lease/safe_deltas"},
		{"a safe state read from an argument", func(d map[string]any) {
			put(d, "rate", "kinds", 2, "effectors", 1, "lease", "safe_deltas", 0, "from_arg")
		}, "/kinds/2/effectors/1/lease/safe_deltas/0"},
		{"a responder delta reading an undeclared argument", func(d map[string]any) {
			put(d, "hours", "kinds", 2, "effectors", 2, "responder", "on_miss", 0, "from_arg")
		}, "/kinds/2/effectors/2/responder/on_miss/0/from_arg"},
		{"a pulse without a duration", func(d map[string]any) { delete(at(d, "kinds", 1, "faults", 2, "onset").(map[string]any), "duration_s") }, "/kinds/1/faults/2/onset/duration_s"},
		{"a duration on a step fault", func(d map[string]any) { put(d, 2.0, "kinds", 1, "faults", 0, "onset", "duration_s") }, "/kinds/1/faults/0/onset/duration_s"},
		{"a revert lease of an assignment", func(d map[string]any) { put(d, "revert", "kinds", 2, "effectors", 1, "lease", "fallback") }, "/kinds/2/effectors/1/lease/fallback"},
	})
}
