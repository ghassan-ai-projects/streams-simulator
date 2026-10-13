package domain

import (
	"strings"
	"testing"
)

func changing(changes map[string]any) map[string]any {
	return kilnEntry(map[string]any{"instruments": changes})
}

func channelNames(t *testing.T, doc map[string]any) []string {
	t.Helper()
	var names []string
	for _, channel := range member(t, doc, "kinds", 0, "channels").([]any) {
		names = append(names, channel.(map[string]any)["name"].(string))
	}
	return names
}

func TestComposeAddsAnOptionalPointAndRecordsItsInstrument(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, changing(map[string]any{"add": []any{"heat_flow"}})))
	names := strings.Join(channelNames(t, doc), ",")
	if names != "kiln.temp,kiln.power,kiln.heat_flow" {
		t.Fatalf("channels = %s, want the optional point attached last, in declaration order", names)
	}
	if got := member(t, doc, "instances", 0, "provenance", "instruments", "heat_flow"); got != "acme/meter@1" {
		t.Errorf("provenance records %v for heat_flow, want acme/meter@1", got)
	}
}

func TestComposeRemovesAPointTogetherWithItsChannelAndFaults(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, changing(map[string]any{"remove": []any{"power"}})))
	if names := strings.Join(channelNames(t, doc), ","); names != "kiln.temp" {
		t.Fatalf("channels = %s, want only kiln.temp", names)
	}
	for _, fault := range member(t, doc, "kinds", 0, "faults").([]any) {
		if strings.Contains(fault.(map[string]any)["id"].(string), "power") {
			t.Errorf("fault %v of the removed point survived", fault)
		}
	}
}

func TestComposeReplacesTheInstrumentAndRecordsItsDigest(t *testing.T) {
	t.Parallel()
	variant := instrumentVariant(t, "acme/probe-k", "K", func(doc map[string]any) {})
	library := libraryWith(t, variant)
	doc := composeOK(t, library, withEquipment(t, replacing("temp", "acme/probe-k@1")))
	provenance := member(t, doc, "instances", 0, "provenance")
	if member(t, provenance, "instruments", "temp") != "acme/probe-k@1" {
		t.Errorf("instruments = %v, want the replacement recorded", member(t, provenance, "instruments"))
	}
	want := library.instruments["acme/probe-k@1"].Digest
	if got := member(t, provenance, "instrument_digests", "acme/probe-k@1"); got != want {
		t.Errorf("replacement digest = %v, want %s", got, want)
	}
}

func TestComposeRefusesInstrumentChangesThatDoNotFit(t *testing.T) {
	t.Parallel()
	library := libraryWith(t, instrumentVariant(t, "acme/flow", "degC", func(doc map[string]any) {
		doc["tag"], doc["unit_family"], doc["unit"] = "JT", "power", "kW"
	}))
	cases := []struct {
		name, pointer, message string
		changes                map[string]any
	}{
		{"add a point that is not optional", "/equipment/0/instruments/add/0", "is not optional", map[string]any{"add": []any{"temp"}}},
		{"add an unknown point", "/equipment/0/instruments/add/0", "has no measurement point", map[string]any{"add": []any{"flow"}}},
		{"remove a point that is not declared", "/equipment/0/instruments/remove/0", "is not attached", map[string]any{"remove": []any{"flow"}}},
		{"remove an optional point that was never added", "/equipment/0/instruments/remove/0", "is not attached", map[string]any{"remove": []any{"heat_flow"}}},
		{"replace a point that is not attached", "/equipment/0/instruments/replace/heat_flow", "is not attached", map[string]any{"replace": map[string]any{"heat_flow": "acme/meter@1"}}},
		{"replace a removed point", "/equipment/0/instruments/replace/power", "is not attached", map[string]any{"remove": []any{"power"}, "replace": map[string]any{"power": "acme/meter@1"}}},
		{"replace with an instrument the library lacks", "/equipment/0/instruments/replace/temp", "is not an instrument of the library", map[string]any{"replace": map[string]any{"temp": "acme/nothing@1"}}},
		{"replace with an instrument of another tag", "/equipment/0/instruments/replace/temp", "measures JT/power", map[string]any{"replace": map[string]any{"temp": "acme/flow@1"}}},
		{"remove the point a fault detector needs", "/equipment/0/instruments", `the detector of fault fouling needs measurement point "temp"`, map[string]any{"remove": []any{"temp"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, library, withEquipment(t, changing(tc.changes)))
			wantIssue(t, err, "composition-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestComposeRefusesRemovingAPointAnEffectorConfirmationNeeds(t *testing.T) {
	t.Parallel()
	confirmed := mutate(t, kilnJSON, func(doc map[string]any) {
		effectors := member(t, doc, "kind", "effectors").([]any)
		effectors[0].(map[string]any)["confirmation_channels"] = []any{"power"}
	})
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", confirmed)}
	library := loadLibrary(t, files.directory("lib"))
	_, err := compose(t, library, withEquipment(t, changing(map[string]any{"remove": []any{"power"}})))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/instruments", "a confirmation of effector set_setpoint needs")
}

func TestComposeRefusesAnOptionalPointAFaultNeedsWhenItIsNotAdded(t *testing.T) {
	t.Parallel()
	needing := mutate(t, kilnJSON, func(doc map[string]any) {
		faults := member(t, doc, "kind", "faults").([]any)
		detector := member(t, faults[0], "observability", "detector").(map[string]any)
		detector["channel"] = "heat_flow"
	})
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", needing)}
	library := loadLibrary(t, files.directory("lib"))
	_, err := compose(t, library, withEquipment(t, kilnEntry(nil)))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/instruments", `needs measurement point "heat_flow"`)
	composeOK(t, library, withEquipment(t, changing(map[string]any{"add": []any{"heat_flow"}})))
}
