package domain

import "testing"

func TestNarrowingMayOnlyTightenWhatTheKindDeclares(t *testing.T) {
	t.Parallel()
	bounded := func(d map[string]any) {
		put(d, 0.0, "kinds", 0, "state", 0, "min")
		put(d, 5.0, "kinds", 0, "state", 0, "max")
	}
	narrowLevel := func(lo, hi float64) func(d map[string]any) {
		return func(d map[string]any) {
			bounded(d)
			put(d, lo, "instances", 1, "override", "narrow", "state", 0, "min")
			put(d, hi, "instances", 1, "override", "narrow", "state", 0, "max")
		}
	}
	runViolations(t, []violationCase{
		{"a state max that widens", narrowLevel(1, 6), "/instances/1/override/narrow/state/0/max"},
		{"a state min that widens", narrowLevel(-1, 4), "/instances/1/override/narrow/state/0/min"},
		{"a state min above its max", narrowLevel(4, 3), "/instances/1/override/narrow/state/0/max"},
		{"bounds that exclude the initial value", narrowLevel(2.5, 4), "/instances/1/override/narrow/state/0/min"},
		{"a state the kind does not declare", func(d map[string]any) {
			put(d, "nothing", "instances", 1, "override", "narrow", "state", 0, "name")
		}, "/instances/1/override/narrow/state/0/name"},
		{"a channel range that widens", func(d map[string]any) {
			put(d, 25.0, "instances", 3, "override", "narrow", "channels", 0, "range", "max")
		}, "/instances/3/override/narrow/channels/0/range/max"},
		{"a channel range that widens below", func(d map[string]any) {
			put(d, -1.0, "instances", 3, "override", "narrow", "channels", 0, "range", "min")
		}, "/instances/3/override/narrow/channels/0/range/min"},
		{"a channel the kind does not declare", func(d map[string]any) {
			put(d, "pump.rpm", "instances", 3, "override", "narrow", "channels", 0, "name")
		}, "/instances/3/override/narrow/channels/0/name"},
		{"an argument maximum that widens", func(d map[string]any) {
			put(d, 150.0, "instances", 3, "override", "narrow", "effectors", 0, "args", "speed", "maximum")
		}, "/instances/3/override/narrow/effectors/0/args/speed/maximum"},
		{"an argument minimum that widens", func(d map[string]any) {
			put(d, map[string]any{"minimum": -5.0}, "instances", 3, "override", "narrow", "effectors", 0, "args", "speed")
		}, "/instances/3/override/narrow/effectors/0/args/speed/minimum"},
		{"an argument with minimum above maximum", func(d map[string]any) {
			put(d, map[string]any{"minimum": 40.0, "maximum": 30.0}, "instances", 3, "override", "narrow", "effectors", 0, "args", "speed")
		}, "/instances/3/override/narrow/effectors/0/args/speed/maximum"},
		{"an argument the schema does not declare", func(d map[string]any) {
			put(d, map[string]any{"maximum": 1.0}, "instances", 3, "override", "narrow", "effectors", 0, "args", "nope")
			delete(at(d, "instances", 3, "override", "narrow", "effectors", 0, "args").(map[string]any), "speed")
		}, "/instances/3/override/narrow/effectors/0/args/nope"},
		{"an argument that is not a number", func(d map[string]any) {
			put(d, map[string]any{"maximum": 1.0}, "instances", 3, "override", "narrow", "effectors", 0, "args", "pump_id")
			delete(at(d, "instances", 3, "override", "narrow", "effectors", 0, "args").(map[string]any), "speed")
		}, "/instances/3/override/narrow/effectors/0/args/pump_id"},
		{"an effector the kind does not declare", func(d map[string]any) {
			put(d, "nothing", "instances", 3, "override", "narrow", "effectors", 0, "name")
		}, "/instances/3/override/narrow/effectors/0/name"},
	})
}

func TestNarrowingWithinTheDeclaredBoundsIsAccepted(t *testing.T) {
	t.Parallel()
	for _, hi := range []float64{4, 5} {
		doc := structuredDocument(t, func(d map[string]any) {
			put(d, 0.0, "kinds", 0, "state", 0, "min")
			put(d, 5.0, "kinds", 0, "state", 0, "max")
			put(d, hi, "instances", 1, "override", "narrow", "state", 0, "max")
		})
		if _, err := Parse(doc, "structured.json"); err != nil {
			t.Fatalf("max %v: %v", hi, err)
		}
	}
}
