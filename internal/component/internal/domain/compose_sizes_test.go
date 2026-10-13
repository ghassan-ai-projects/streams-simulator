package domain

import (
	"testing"
)

func TestComposeAppliesThePresetThenTheParameters(t *testing.T) {
	t.Parallel()
	entry := kilnEntry(map[string]any{"preset": "big", "parameters": map[string]any{"setpoint_c": 120, "insulation": "poor", "stages": 3}})
	doc := composeOK(t, baseLibrary(t), withEquipment(t, entry))
	states := member(t, doc, "kinds", 0, "state")
	checks := []struct {
		state, field string
		want         float64
	}{{"load", "initial", 500}, {"setpoint", "initial", 120}, {"temp", "initial", 120}, {"load", "min", 3}}
	for _, tc := range checks {
		if got := number(t, named(t, states, "state", "name", tc.state)[tc.field]); got != tc.want {
			t.Errorf("state %s %s = %v, want %v", tc.state, tc.field, got, tc.want)
		}
	}
	lag := member(t, named(t, member(t, doc, "kinds", 0, "dynamics"), "dynamics", "target", "temp"), "f1", "time_constant_s")
	if number(t, lag) != 120 {
		t.Errorf("time constant = %v, want the poor insulation case 120", lag)
	}
	if got := number(t, member(t, doc, "kinds", 0, "channels", 1, "range", "max")); got != 1000 {
		t.Errorf("power range max = %v, want the preset's 1000", got)
	}
}

func TestComposeUsesTheDefaultPresetAndTheParameterDefaults(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	provenance := member(t, doc, "instances", 0, "provenance")
	if got := member(t, provenance, "preset"); got != "std" {
		t.Errorf("preset = %v, want the default std", got)
	}
	parameters := member(t, provenance, "parameters").(map[string]any)
	if number(t, parameters["setpoint_c"]) != 100 || parameters["insulation"] != "good" || number(t, parameters["stages"]) != 0 {
		t.Errorf("parameters = %v, want the defaults recorded", parameters)
	}
}

func TestComposeRefusesAPresetTheComponentDoesNotHave(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"preset": "huge"})))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/preset", `has no preset "huge"`)
}

func TestComposeRefusesParametersThatBreakTheirDeclaration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		parameters             map[string]any
	}{
		{"unknown name", "/equipment/0/parameters/shade", `has no parameter "shade"`, map[string]any{"shade": 1}},
		{"below the minimum", "/equipment/0/parameters/setpoint_c", "outside the bounds", map[string]any{"setpoint_c": 10}},
		{"above the maximum", "/equipment/0/parameters/setpoint_c", "outside the bounds", map[string]any{"setpoint_c": 151}},
		{"a string for a number", "/equipment/0/parameters/setpoint_c", "takes a number", map[string]any{"setpoint_c": "hot"}},
		{"a fraction for an integer", "/equipment/0/parameters/stages", "takes a whole number", map[string]any{"stages": 1.5}},
		{"a value outside an enum", "/equipment/0/parameters/insulation", "is not one of", map[string]any{"insulation": "foam"}},
		{"a number for an enum", "/equipment/0/parameters/insulation", "is not one of", map[string]any{"insulation": 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"parameters": tc.parameters})))
			wantIssue(t, err, "composition-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestComposeAcceptsTheBoundsThemselves(t *testing.T) {
	t.Parallel()
	for _, value := range []int{50, 150} {
		entry := kilnEntry(map[string]any{"parameters": map[string]any{"setpoint_c": value}})
		composeOK(t, baseLibrary(t), withEquipment(t, entry))
	}
}
