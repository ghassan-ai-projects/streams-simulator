package domain

import (
	"strings"
	"testing"
)

func loadKiln(t *testing.T, change func(doc map[string]any)) error {
	t.Helper()
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", mutate(t, kilnJSON, change))}
	_, err := LoadLibrary([]Directory{files.directory("lib")})
	return err
}

func list(t *testing.T, doc any, path ...any) []any {
	t.Helper()
	items, _ := member(t, doc, path...).([]any)
	return items
}

func object(t *testing.T, doc any, path ...any) map[string]any {
	t.Helper()
	items, _ := member(t, doc, path...).(map[string]any)
	return items
}

func TestLoadRefusesPresetsAndParametersThatNameNothing(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a preset member that does not exist", "/presets/big/values/state~1setpoint~1shade", `no member "shade"`,
			func(doc map[string]any) { object(t, doc, "presets", "big", "values")["state/setpoint/shade"] = 1 }},
		{"a preset state that does not exist", "/presets/big/values/state~1nothing~1initial", `state has no entry "nothing"`,
			func(doc map[string]any) { object(t, doc, "presets", "big", "values")["state/nothing/initial"] = 1 }},
		{"a preset value of another type", "/presets/big/values/state~1setpoint~1initial", "holds a number, not a string",
			func(doc map[string]any) { object(t, doc, "presets", "big", "values")["state/setpoint/initial"] = "hot" }},
		{"a position beyond the list", "/presets/big/values/dynamics~1load~1f0~1seasonality~15~1amplitude", "outside the 1 items",
			func(doc map[string]any) {
				object(t, doc, "presets", "big", "values")["dynamics/load/f0/seasonality/5/amplitude"] = 1
			}},
		{"a parameter target that does not exist", "/parameters/0/targets", `state has no entry "nothing"`,
			func(doc map[string]any) {
				first := list(t, doc, "parameters")[0].(map[string]any)
				first["targets"] = append(first["targets"].([]any), "state/nothing/initial")
			}},
		{"a case target that does not exist", "/parameters/1/cases/good", `no member "nothing"`,
			func(doc map[string]any) {
				object(t, doc, "parameters", 1, "cases", "good")["dynamics/temp/f1/nothing"] = 1
			}},
		{"a target written by the preset and a parameter", "/presets/big/values/state~1setpoint~1initial", `also written by parameter "setpoint_c"`,
			func(doc map[string]any) { object(t, doc, "presets", "big", "values")["state/setpoint/initial"] = 5 }},
		{"a default preset that is no preset", "/default_preset", `default preset "huge"`,
			func(doc map[string]any) { doc["default_preset"] = "huge" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantIssue(t, loadKiln(t, tc.change), "component-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesParametersThatBreakTheirOwnDeclaration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a default above the maximum", "/parameters/0/default", "outside the bounds", func(doc map[string]any) { list(t, doc, "parameters")[0].(map[string]any)["default"] = 500 }},
		{"a default below the minimum", "/parameters/0/default", "outside the bounds", func(doc map[string]any) { list(t, doc, "parameters")[0].(map[string]any)["default"] = 10 }},
		{"a default that is not a number", "/parameters/0/default", "must be a number", func(doc map[string]any) { list(t, doc, "parameters")[0].(map[string]any)["default"] = "hot" }},
		{"a fractional integer default", "/parameters/2/default", "is not whole", func(doc map[string]any) { list(t, doc, "parameters")[2].(map[string]any)["default"] = 1.5 }},
		{"an enum default outside its values", "/parameters/1/default", "is not one of the values", func(doc map[string]any) { list(t, doc, "parameters")[1].(map[string]any)["default"] = "foam" }},
		{"cases that differ from the values", "/parameters/1/cases", "differ from values", func(doc map[string]any) { delete(object(t, doc, "parameters", 1, "cases"), "poor") }},
		{"two parameters writing one target", "/parameters/2", `target "state/setpoint/initial" is also written by parameter "setpoint_c"`, func(doc map[string]any) {
			list(t, doc, "parameters")[2].(map[string]any)["targets"] = []any{"state/setpoint/initial"}
		}},
		{"a parameter writing a target an enum case writes", "/parameters/2", `target "dynamics/temp/f1/time_constant_s" is also written by parameter "insulation"`, func(doc map[string]any) {
			list(t, doc, "parameters")[2].(map[string]any)["targets"] = []any{"dynamics/temp/f1/time_constant_s"}
		}},
		{"a parameter declared twice", "/parameters/1/name", `parameter "setpoint_c" is declared twice`, func(doc map[string]any) { list(t, doc, "parameters")[1].(map[string]any)["name"] = "setpoint_c" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantIssue(t, loadKiln(t, tc.change), "component-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesPointsAndPortsThatAreNotWhatTheyClaim(t *testing.T) {
	t.Parallel()
	point := func(doc map[string]any, i int) map[string]any {
		return list(t, doc, "measurement_points")[i].(map[string]any)
	}
	port := func(doc map[string]any, i int) map[string]any { return list(t, doc, "ports")[i].(map[string]any) }
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a point on an undeclared state", "/measurement_points/0/state", `state "nothing" is not declared`, func(doc map[string]any) { point(doc, 0)["state"] = "nothing" }},
		{"a point whose family is not in the vocabulary", "/measurement_points/0/unit_family", `unit family "gizmo" is not in the vocabulary`, func(doc map[string]any) { point(doc, 0)["unit_family"] = "gizmo" }},
		{"a state unit outside the point family", "/measurement_points/0/state", `has unit "degC", which is not a unit of family "power"`, func(doc map[string]any) { point(doc, 0)["unit_family"] = "power" }},
		{"a tag that is not in the vocabulary", "/measurement_points/0/tag", `tag "ZZ" is not in the vocabulary`, func(doc map[string]any) { point(doc, 0)["tag"] = "ZZ" }},
		{"a tag that does not measure the family", "/measurement_points/1/tag", `tag "TT" does not measure unit family "power"`, func(doc map[string]any) { point(doc, 1)["tag"] = "TT" }},
		{"an inverted range", "/measurement_points/0/range", "must be below max", func(doc map[string]any) { point(doc, 0)["range"] = map[string]any{"min": 5, "max": 5} }},
		{"a default instrument the library lacks", "/measurement_points/0/default_instrument", "acme/nothing@1 is not in the library", func(doc map[string]any) { point(doc, 0)["default_instrument"] = "acme/nothing@1" }},
		{"a default instrument of another tag", "/measurement_points/0/default_instrument", "measures JT/power, the point TT/temperature", func(doc map[string]any) { point(doc, 0)["default_instrument"] = "acme/meter@1" }},
		{"a point declared twice", "/measurement_points/1/name", `measurement point "temp" is declared twice`, func(doc map[string]any) { point(doc, 1)["name"] = "temp" }},
		{"a port declared twice", "/ports/1/name", `port "load_in" is declared twice`, func(doc map[string]any) { port(doc, 1)["name"] = "load_in" }},
		{"a port type that is not in the vocabulary", "/ports/0/type", `port type "steam" is not in the vocabulary`, func(doc map[string]any) { port(doc, 0)["type"] = "steam" }},
		{"a port on an undeclared state", "/ports/0/state", `state "nothing" is not declared`, func(doc map[string]any) { port(doc, 0)["state"] = "nothing" }},
		{"a port whose state unit is outside the type family", "/ports/0/state", `fixed by port type "thermal_load"`, func(doc map[string]any) { port(doc, 0)["state"] = "temp" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantIssue(t, loadKiln(t, tc.change), "component-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesReferencesToMeasurementPointsThatDoNotExist(t *testing.T) {
	t.Parallel()
	extra := map[string]any{"id": "frozen", "observation": map[string]any{"channel": "nothing", "form": "freeze"},
		"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": "temp"}}, "onset": map[string]any{"shape": "step"}}
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a detector channel", "/kind/faults/0/observability/detector/channel", `names "nothing", which is not a measurement point`,
			func(doc map[string]any) {
				object(t, doc, "kind", "faults", 0, "observability", "detector")["channel"] = "nothing"
			}},
		{"an observation channel", "/kind/faults/1/observation/channel", "the observation of fault frozen",
			func(doc map[string]any) {
				object(t, doc, "kind")["faults"] = append(list(t, doc, "kind", "faults"), extra)
			}},
		{"a confirmation channel", "/kind/effectors/0/confirmation_channels/0", "a confirmation of effector set_setpoint",
			func(doc map[string]any) {
				list(t, doc, "kind", "effectors")[0].(map[string]any)["confirmation_channels"] = []any{"nothing"}
			}},
		{"an expected effector", "/kind/faults/0/expected_effector", `expected effector "nothing"`,
			func(doc map[string]any) {
				list(t, doc, "kind", "faults")[0].(map[string]any)["expected_effector"] = "nothing"
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantIssue(t, loadKiln(t, tc.change), "component-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesNamesThatBreakTheQualifiedNameRules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a fault with a double underscore", "/kind/faults/0/id", "must match", func(doc map[string]any) { list(t, doc, "kind", "faults")[0].(map[string]any)["id"] = "bad__id" }},
		{"an effector beyond forty characters", "/kind/effectors/0/name", "at most 40 characters",
			func(doc map[string]any) {
				list(t, doc, "kind", "effectors")[0].(map[string]any)["name"] = strings.Repeat("a", 41)
			}},
		{"a component whose kind name has a double underscore", "/id", "must match", func(doc map[string]any) { doc["id"] = "acme/bad--name" }},
		{"a kind section holding a non-object", "/kind/faults/0", "must be an object", func(doc map[string]any) { object(t, doc, "kind")["faults"] = []any{1} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantIssue(t, loadKiln(t, tc.change), "component-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesAFileThatBreaksItsSchema(t *testing.T) {
	t.Parallel()
	err := loadKiln(t, func(doc map[string]any) { delete(doc, "title") })
	invalid := asViolation(t, err, "component-v0.1")
	if invalid.Source != "lib/equipment/kiln.component.json" || len(invalid.Issues) == 0 {
		t.Fatalf("error = %+v, want the schema issues naming the file", invalid)
	}
}

func TestLoadRefusesAFileThatIsNotJSON(t *testing.T) {
	t.Parallel()
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", "{")}
	_, err := LoadLibrary([]Directory{files.directory("lib")})
	wantViolation(t, err, "component-v0.1")
}

func TestLoadReportsEveryIssueOfAFile(t *testing.T) {
	t.Parallel()
	err := loadKiln(t, func(doc map[string]any) {
		doc["default_preset"] = "huge"
		list(t, doc, "measurement_points")[0].(map[string]any)["tag"] = "ZZ"
	})
	pointers := map[string]bool{}
	for _, issue := range asViolation(t, err, "component-v0.1").Issues {
		pointers[issue.Pointer] = true
	}
	if !pointers["/default_preset"] || !pointers["/measurement_points/0/tag"] {
		t.Fatalf("issue pointers = %v, want both the preset and the tag problems", pointers)
	}
}
