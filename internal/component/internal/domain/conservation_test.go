package domain

import "testing"

func balancedKiln(t *testing.T, inputs, outputs []any) File {
	t.Helper()
	raw := mutate(t, kilnJSON, func(doc map[string]any) {
		faults := member(t, doc, "kind", "faults").([]any)
		doc["kind"].(map[string]any)["faults"] = append(faults, map[string]any{"id": "imbalance",
			"affects":       []any{map[string]any{"state": "temp", "delta": 2}},
			"observability": map[string]any{"detector": map[string]any{"form": "conservation_residual", "inputs": inputs, "outputs": outputs}, "first_observable_snr": 3, "unavoidable_snr": 10},
			"onset":         map[string]any{"magnitude": 1, "shape": "step"}})
	})
	return file("equipment/kiln.component.json", raw)
}

func balancedLibrary(t *testing.T, inputs, outputs []any) (*Library, error) {
	t.Helper()
	files := baseFiles()
	files.equipment = []File{balancedKiln(t, inputs, outputs)}
	return LoadLibrary([]Directory{files.directory("lib")})
}

func TestLoadRefusesConservationChannelsThatAreNoMeasurementPoint(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		pointer         string
		inputs, outputs []any
	}{
		"an input":  {"/kind/faults/1/observability/detector/inputs/1", []any{"power", "nothing"}, []any{"heat_flow"}},
		"an output": {"/kind/faults/1/observability/detector/outputs/0", []any{"power"}, []any{"nothing"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := balancedLibrary(t, tc.inputs, tc.outputs)
			wantIssue(t, err, "component-v0.1", tc.pointer, `the detector of fault imbalance names "nothing", which is not a measurement point`)
		})
	}
}

func TestComposeQualifiesTheConservationChannelsWithTheKindName(t *testing.T) {
	t.Parallel()
	library, err := balancedLibrary(t, []any{"power"}, []any{"heat_flow"})
	if err != nil {
		t.Fatal(err)
	}
	entry := kilnEntry(map[string]any{"as": "kiln_a", "instruments": map[string]any{"add": []any{"heat_flow"}}})
	doc := composeOK(t, library, withEquipment(t, entry))
	detector := member(t, named(t, member(t, doc, "kinds", 0, "faults"), "fault", "id", "kiln_a__imbalance"), "observability", "detector")
	if member(t, detector, "inputs", 0) != "kiln_a.power" || member(t, detector, "outputs", 0) != "kiln_a.heat_flow" {
		t.Fatalf("detector = %v, want inputs and outputs qualified with the kind name", detector)
	}
}

func TestComposeRefusesAnOptionalPointAConservationDetectorNeedsWhenItIsNotAdded(t *testing.T) {
	t.Parallel()
	library, err := balancedLibrary(t, []any{"power"}, []any{"heat_flow"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = compose(t, library, withEquipment(t, kilnEntry(nil)))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/instruments", `the detector of fault imbalance needs measurement point "heat_flow"`)
}

func TestComposeRefusesRemovingAPointAConservationInputNeeds(t *testing.T) {
	t.Parallel()
	library, err := balancedLibrary(t, []any{"power"}, []any{"heat_flow"})
	if err != nil {
		t.Fatal(err)
	}
	entry := kilnEntry(map[string]any{"instruments": map[string]any{"add": []any{"heat_flow"}, "remove": []any{"power"}}})
	_, err = compose(t, library, withEquipment(t, entry))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/instruments", `the detector of fault imbalance needs measurement point "power"`)
}
