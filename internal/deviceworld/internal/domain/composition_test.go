package domain

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestSelectEntityDefaultsToTheFirstEntityAndRefusesAnUnknownOne(t *testing.T) {
	t.Parallel()
	w := coldChainWorld(t, 1)
	first, err := SelectEntity(w, "")
	if err != nil || first != w.EntityIDs()[0] {
		t.Fatalf("default entity = %q, %v; want %q", first, err, w.EntityIDs()[0])
	}
	if got, err := SelectEntity(w, first); err != nil || got != first {
		t.Fatalf("named entity = %q, %v", got, err)
	}
	if _, err := SelectEntity(w, "no-such"); err == nil || err.Error() != `device world entity "no-such" does not exist` {
		t.Fatalf("an unknown entity: err = %v", err)
	}
}

func TestValidateCompositionRequiresTheSafeStopsTheCatalogDeclares(t *testing.T) {
	t.Parallel()
	w := coldChainWorld(t, 1)
	data, err := os.ReadFile("../../testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(data, w.EntityIDs()[0])
	if err != nil {
		t.Fatal(err)
	}
	caps := thermalCapabilities(t)
	if err := ValidateComposition(w, bindings, caps); err != nil {
		t.Fatalf("the shipped composition must validate: %v", err)
	}
	if stops := safeStops(bindings, caps); !slices.Equal(stops, slices.Sorted(slices.Values(stops))) || len(stops) == 0 {
		t.Fatalf("safe stops = %v, want the bound targets the catalog stops", stops)
	}
	fan := bindings["fan-01"]
	fan.effector = "no_such_effector"
	bindings["fan-01"] = fan
	if err := ValidateComposition(w, bindings, caps); err == nil || !strings.HasPrefix(err.Error(), "validate device world bindings: ") {
		t.Fatalf("a binding to an effector the world lacks: err = %v", err)
	}
}

func thermalCapabilities(t *testing.T) *device.Capabilities {
	t.Helper()
	data, err := os.ReadFile(testsupport.DeviceCapabilities())
	if err != nil {
		t.Fatal(err)
	}
	caps, err := device.LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}
