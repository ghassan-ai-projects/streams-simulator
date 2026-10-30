package domain_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func readDocument(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func normalised(document map[string]any, edit func(kind map[string]any)) map[string]any {
	for _, member := range []string{"id", "title", "description"} {
		delete(document, member)
	}
	edit(document["kinds"].([]any)[0].(map[string]any))
	return document
}

func channelNamed(kind map[string]any, name string) map[string]any {
	for _, channel := range kind["channels"].([]any) {
		if channel.(map[string]any)["name"] == name {
			return channel.(map[string]any)
		}
	}
	return nil
}

func TestThePumpStationWorldsDifferFromTheReferenceOnlyWhereTheyAreMeantTo(t *testing.T) {
	t.Parallel()
	dropMaintenance := func(kind map[string]any) {
		for _, effector := range kind["effectors"].([]any) {
			if effector.(map[string]any)["name"] == "schedule_maintenance" {
				delete(effector.(map[string]any), "responder")
				delete(effector.(map[string]any), "description")
			}
		}
	}
	reference := normalised(readDocument(t, testsupport.Domain("rotating-machinery")), dropMaintenance)
	station := normalised(readDocument(t, testsupport.PumpStationDomain()), dropMaintenance)
	if !reflect.DeepEqual(reference, station) {
		t.Fatal("pump-station differs from rotating-machinery beyond the maintenance responder")
	}
	dropProbe := func(kind map[string]any) {
		dropMaintenance(kind)
		probe := channelNamed(kind, "vibration")
		for _, member := range []string{"range", "resolution", "observation_gain"} {
			delete(probe, member)
		}
	}
	first := normalised(readDocument(t, testsupport.PumpStationDomain()), dropProbe)
	second := normalised(readDocument(t, testsupport.PumpStationProbeDomain()), dropProbe)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("pump-station-probe-b differs from pump-station beyond the vibration probe")
	}
}
