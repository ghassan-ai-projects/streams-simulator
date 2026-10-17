package component

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

const sensorKindJSON = `{
  "kind": "thermal_sensor",
  "metrics": [
    {"name": "temperature", "data_type": "number", "unit": "Cel", "eu_range": {"min": 0, "max": 50}, "resolution": 1, "writable": false},
    {"name": "humidity", "data_type": "number", "unit": "%", "eu_range": {"min": 20, "max": 90}, "resolution": 1, "writable": false}
  ],
  "commands": []
}`

const sensorChoicesJSON = `{
  "schema_version": "0.1", "id": "sensor-twin", "version": "0.1.0", "title": "Sensor twin",
  "kinds": [{"kind": "thermal_sensor", "instance": "zone-01", "metrics": {
    "temperature": {"instrument": "streamsim/thermistor", "initial": 34},
    "humidity": {"instrument": "streamsim/humidity", "initial": 42}}}]
}`

func sensorKinds(t *testing.T) map[string]map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(sensorKindJSON), &doc); err != nil {
		t.Fatal(err)
	}
	return map[string]map[string]any{"thermal_sensor": doc}
}

func TestTwinScaffoldsAKindFileOverBuiltInInstrumentsOfAnotherUnitSpelling(t *testing.T) {
	t.Parallel()
	twin, err := Twin(builtInLibrary(t), sensorKinds(t), []byte(sensorChoicesJSON), "sensor.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(twin.Implements, []string{"thermal_sensor"}) || twin.Summary.Channels != 2 || twin.Summary.Faults < 2 {
		t.Fatalf("twin implements %v with summary %+v, want the sensor kind with two channels and the instruments' faults", twin.Implements, twin.Summary)
	}
}

func TestTwinNeedsALibrary(t *testing.T) {
	t.Parallel()
	for name, library := range map[string]*Library{"no library": nil, "an empty library": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Twin(library, sensorKinds(t), []byte(sensorChoicesJSON), "sensor.twin.json"); !errors.Is(err, ErrNoLibrary) {
				t.Fatalf("Twin = %v, want ErrNoLibrary", err)
			}
		})
	}
}
