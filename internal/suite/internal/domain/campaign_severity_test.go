package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestASeverityRangeDrawsOnlySeveritiesAndMovesNoOtherDraw(t *testing.T) {
	t.Parallel()
	plain := CampaignPhase{Name: "many", Domain: "pump-station", Runs: 8, Faults: threeFaults(), FaultsPerRun: intPtr(2), NegativesPerRun: 1}
	ranged := plain
	ranged.Faults = threeFaults()
	ranged.Faults["misalignment"] = FaultWeight{Weight: 1, SeverityRange: []float64{0.5, 0.7}}
	before, after := buildPlans(t, multiSpec(plain)), buildPlans(t, multiSpec(ranged))
	again := buildPlans(t, multiSpec(ranged))
	if !reflect.DeepEqual(after, again) {
		t.Fatal("two builds of one spec with a severity range differ")
	}
	drawn := 0
	for i := range after {
		for j, d := range after[i].Disturbances {
			if (d.Fault == "misalignment") != (d.Severity != nil) || (d.Severity != nil && (*d.Severity < 0.5 || *d.Severity > 0.7)) {
				t.Fatalf("run %d: %s carries severity %v, want one in [0.5, 0.7] exactly on misalignment", i+1, d.Fault, d.Severity)
			}
			if d.Severity != nil {
				drawn++
			}
			d.Severity = nil
			if !reflect.DeepEqual(d, before[i].Disturbances[j]) || *after[i].Seed != *before[i].Seed {
				t.Fatalf("run %d disturbance %d: %+v with a range, %+v without; the range moved another draw", i+1, j+1, d, before[i].Disturbances[j])
			}
		}
	}
	if drawn == 0 {
		t.Fatal("no misalignment disturbance drew a severity")
	}
}

func TestAFaultWeightReadsAsANumberOrAsTheClosedObject(t *testing.T) {
	t.Parallel()
	var number, object FaultWeight
	if err := number.UnmarshalJSON([]byte(`2.5`)); err != nil || !reflect.DeepEqual(number, FaultWeight{Weight: 2.5}) {
		t.Fatalf("a number reads as %+v, %v", number, err)
	}
	if err := object.UnmarshalJSON([]byte(`{"weight": 3, "severity_range": [1, 2]}`)); err != nil ||
		!reflect.DeepEqual(object, FaultWeight{Weight: 3, SeverityRange: []float64{1, 2}}) {
		t.Fatalf("an object reads as %+v, %v", object, err)
	}
	if err := object.UnmarshalJSON([]byte(`{"weight": 3, "severity": 1}`)); err == nil || !strings.Contains(err.Error(), `unknown field "severity"`) {
		t.Fatalf("an object with an unknown member = %v, want it refused naming the member", err)
	}
}
