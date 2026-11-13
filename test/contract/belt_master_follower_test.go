package contract

import (
	"reflect"
	"slices"
	"testing"
)

var (
	referenceStates     = []string{"ref_seen", "ref_in", "ref_held", "promoted", "speed_lead", "ref_ok"}
	referencePorts      = []string{"reference_in", "reference_out"}
	referenceParameters = []string{"role", "master_loss", "promote_s"}
)

func withoutNamed(list any, key string, names []string) []any {
	var kept []any
	for _, item := range list.([]any) {
		if !slices.Contains(names, item.(map[string]any)[key].(string)) {
			kept = append(kept, item)
		}
	}
	return kept
}

// withoutReferenceLink removes from the master-follower drive what its reference link adds to the conveyor drive.
func withoutReferenceLink(doc map[string]any) map[string]any {
	kind := doc["kind"].(map[string]any)
	kind["state"] = withoutNamed(kind["state"], "name", referenceStates)
	kind["dynamics"] = withoutNamed(kind["dynamics"], "target", referenceStates)
	for _, dynamic := range kind["dynamics"].([]any) {
		if f1, _ := dynamic.(map[string]any)["f1"].(map[string]any); dynamic.(map[string]any)["target"] == "enabled" {
			f1["inputs"] = withoutNamed(f1["inputs"], "state", []string{"ref_ok"})
		}
	}
	doc["ports"] = withoutNamed(doc["ports"], "name", referencePorts)
	doc["parameters"] = withoutNamed(doc["parameters"], "name", referenceParameters)
	for _, member := range []string{"version", "title", "description"} {
		delete(doc, member)
	}
	return doc
}

func TestTheMasterFollowerDriveIsTheConveyorDrivePlusItsReferenceLink(t *testing.T) {
	t.Parallel()
	lib := loadLibrary(t)
	first, second := lib.files["streamsim/conveyor-drive@1"], lib.files["streamsim/conveyor-drive@2"]
	for _, member := range []string{"version", "title", "description"} {
		delete(first, member)
	}
	if !reflect.DeepEqual(withoutReferenceLink(second), first) {
		t.Fatal("conveyor-drive@2 without its reference states, ports and parameters differs from conveyor-drive@1")
	}
}

// Until the master is lost, a master-follower plant is plant a: the same trajectories of every state of plant a, bit for bit,
// through a trip of D-3, the last drive of the line, which no drive follows.
func TestTheMasterFollowerPlantsArePlantAUntilTheMasterIsLost(t *testing.T) {
	t.Parallel()
	reference := newBeltWorld(t, beltPlant(t, "a"))
	plants := map[string]*beltWorld{"s6-stop": newBeltWorld(t, beltPlant(t, "s6-stop")), "s6-promote": newBeltWorld(t, beltPlant(t, "s6-promote"))}
	for _, at := range []float64{300, faultAtS, 900, 1800} {
		reference.advance(at)
		for _, b := range plants {
			b.advance(at)
		}
		if at == faultAtS {
			reference.inject("D-3", "conveyor_drive__vfd_fault", 1)
			for _, b := range plants {
				b.inject("D-3", "drive_follower__vfd_fault", 1)
			}
		}
		for name, b := range plants {
			for _, entity := range reference.w.EntityIDs() {
				for _, state := range reference.w.StateNamesOf(entity) {
					if got, want := b.state(entity, state), reference.state(entity, state); got != want {
						t.Errorf("%s at %.0f s: %s.%s = %v, plant a %v", name, at, entity, state, got, want)
					}
				}
			}
		}
	}
}
