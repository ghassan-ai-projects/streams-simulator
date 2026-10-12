package domain

import (
	"strings"
	"testing"
)

func TestNamedInstancesResolveBeforeTemplatesAndJoinTheirSet(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	cases := []struct {
		id   string
		set  int
		name bool
	}{
		{"tank-1", 0, false}, {"T-main", 1, true}, {"pump-2", 2, false}, {"P-slow", 3, true}, {"air_header", 4, true}, {"T-other", -1, false},
	}
	for _, tc := range cases {
		if got := c.InstanceSetFor(tc.id); got != tc.set || c.IsInstanceName(tc.id) != tc.name {
			t.Fatalf("%s: set %d name %v, want set %d name %v", tc.id, got, c.IsInstanceName(tc.id), tc.set, tc.name)
		}
	}
}

func TestExplicitIdsMayBeNamesAndAreCountedOnlyWhenTemplated(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	if err := c.CheckEntityIDs([]string{"T-main", "P-slow", "air_header", "tank-1"}); err != nil {
		t.Fatalf("names with a templated id: %v", err)
	}
	if err := c.CheckEntityIDs([]string{"T-main"}); err != nil {
		t.Fatalf("a name alone does not reach a count minimum: %v", err)
	}
	if err := c.CheckEntityIDs([]string{"tank-1", "T-other"}); err == nil || !strings.Contains(err.Error(), `"T-other" does not fit`) {
		t.Fatalf("an unknown name: %v, want a refusal naming it", err)
	}
	tooMany := []string{"tank-1", "tank-2", "tank-3", "tank-4", "tank-5", "tank-6", "tank-7"}
	if err := c.CheckEntityIDs(tooMany); err == nil || !strings.Contains(err.Error(), "above the instance count maximum 6") {
		t.Fatalf("seven tanks: %v", err)
	}
}

func TestInstanceSetsOfNamesAreCheckedAtTheirPointers(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"a name used by two sets", func(d map[string]any) { put(d, "T-main", "instances", 3, "names", 0) }, "/instances/3/names/0"},
		{"a name that fits another set's template", func(d map[string]any) { put(d, "pump-4", "instances", 1, "names", 0) }, "/instances/1/names/0"},
		{"a name that fits its own template", func(d map[string]any) { put(d, []any{"tank-9"}, "instances", 0, "names") }, "/instances/0/names/0"},
		{"neither names nor template", func(d map[string]any) {
			put(d, map[string]any{"kind": "pump"}, "instances", 3)
		}, "/instances/3/id_template"},
		{"a template without a count", func(d map[string]any) {
			delete(at(d, "instances", 0).(map[string]any), "count")
		}, "/instances/0/count"},
		{"churn on names only", func(d map[string]any) {
			put(d, map[string]any{"births_per_hour": 1.0}, "instances", 1, "churn")
		}, "/instances/1/churn"},
		{"a count on names only", func(d map[string]any) {
			put(d, map[string]any{"default": 1}, "instances", 4, "count")
		}, "/instances/4/count"},
	})
}
