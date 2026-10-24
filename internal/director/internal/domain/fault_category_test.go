package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestEveryDeclaredFaultIsAWorldFaultAMeasurementFaultOrBoth(t *testing.T) {
	t.Parallel()
	effect := []model.FaultEffect{{State: "temperature", Delta: 1}}
	instrument := &model.ObservationFault{Channel: "temperature", Form: "dropout"}
	spec := &model.DomainSpec{Kinds: []model.Kind{
		{Name: "fan", Faults: []model.Fault{{ID: "blocked", Affects: effect}, {ID: "sensor_dropout", Observation: instrument}}},
		{Name: "pump", Faults: []model.Fault{{ID: "seizes_and_reads_stuck", Affects: effect, Observation: instrument}}},
		{Name: "bare"},
	}}
	want := map[string]map[string]string{
		"fan":  {"blocked": FaultWorld, "sensor_dropout": FaultMeasurement},
		"pump": {"seizes_and_reads_stuck": FaultWorldAndMeasurement},
		"bare": {},
	}
	if got := FaultCategories(spec); !reflect.DeepEqual(got, want) {
		t.Fatalf("categories = %v, want %v", got, want)
	}
}
