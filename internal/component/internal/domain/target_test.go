package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func kilnWorkspace(t *testing.T) *workspace {
	t.Helper()
	library := baseLibrary(t)
	return newWorkspace(library.equipment["acme/kiln@1"])
}

func TestTargetsAddressKeyedAndPositionalItems(t *testing.T) {
	t.Parallel()
	cases := []struct {
		target string
		value  any
	}{
		{"state/setpoint/initial", json.Number("7")},
		{"dynamics/temp/f1/inputs/load/coef", json.Number("0.5")},
		{"dynamics/load/f0/seasonality/0/amplitude", json.Number("99")},
		{"faults/fouling/affects/temp/delta", json.Number("9")},
		{"faults/fouling/onset/shape", "ramp"},
		{"effectors/set_setpoint/ack/failure_modes/ok/probability", json.Number("0.5")},
		{"effectors/set_setpoint/effect/state_deltas/setpoint/assign_from_arg", "other"},
		{"measurement_points/power/range/max", json.Number("5")},
		{"measurement_points/heat_flow/optional", false},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			t.Parallel()
			workspace := kilnWorkspace(t)
			if err := workspace.write(tc.target, tc.value); err != nil {
				t.Fatal(err)
			}
			found, err := workspace.locate(tc.target)
			if err != nil || found.get() != tc.value {
				t.Fatalf("after the write, target holds %v (%v), want %v", found.get(), err, tc.value)
			}
		})
	}
}

func TestTargetsRefuseWhatResolvesToNothing(t *testing.T) {
	t.Parallel()
	cases := []struct{ target, message string }{
		{"state/setpoint", "is not <section>/<name>/<path>"},
		{"state/nothing/initial", `state has no entry "nothing"`},
		{"state/setpoint/shade", `no member "shade"`},
		{"dynamics/temp/f1/inputs/nothing/coef", `no item with state "nothing"`},
		{"dynamics/load/f0/seasonality/3/amplitude", "position \"3\" is outside the 1 items"},
		{"dynamics/load/f0/seasonality/first/amplitude", "position \"first\" is outside"},
		{"dynamics/load/f0", "is an object or list"},
		{"dynamics/load/f0/seasonality", "is an object or list"},
		{"state/setpoint/initial/deeper", "is below a value that has no members"},
		{"flows/pump/rate", `flows has no entry "pump"`},
		{"measurement_points/nothing/range", `measurement_points has no entry "nothing"`},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			t.Parallel()
			_, err := kilnWorkspace(t).locate(tc.target)
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("locate(%s) error = %v, want it to contain %q", tc.target, err, tc.message)
			}
		})
	}
}

func TestTargetsRefuseAValueOfAnotherType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		target string
		value  any
		want   string
	}{
		{"state/setpoint/initial", "hot", "holds a number, not a string"},
		{"state/setpoint/unit", json.Number("1"), "holds a string, not a number"},
		{"measurement_points/heat_flow/optional", "no", "holds a boolean, not a string"},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			t.Parallel()
			err := kilnWorkspace(t).write(tc.target, tc.value)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("write error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestAWorkspaceIsACopyOfTheLibraryFile(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	equipment := library.equipment["acme/kiln@1"]
	workspace := newWorkspace(equipment)
	if err := workspace.write("state/setpoint/initial", json.Number("1")); err != nil {
		t.Fatal(err)
	}
	original := member(t, equipment.Document, "kind", "state", 1, "initial")
	if original != json.Number("100") {
		t.Fatalf("the library file now holds %v, want the untouched 100", original)
	}
}
