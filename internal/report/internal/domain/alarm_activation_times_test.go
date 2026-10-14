package domain

import (
	"encoding/json"
	"slices"
	"testing"
)

const activationTruth = `{"stream":"director","seq":1,"type":"kpi.truth","at_ns":5000,"kpis":{"from_ns":1000,"to_ns":5000,"entities":[{"entity_id":"a","alarms":[{"state":"hot","priority":"high","activations":2,"activation_times_ns":[2000,4000],"chattering":false,"active_s":1,"longest_active_s":1,"active_at_end":true}]}],"alarm_system":{"activations":2}}}`

func TestReportForwardsAlarmActivationInstantsWithoutReconstructingThem(t *testing.T) {
	t.Parallel()
	report, err := Assemble(Folder{Observer: doc(openedExplore), Director: doc(activationTruth)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	alarm := report.Truth.KPIs.Entities[0].Alarms[0]
	if alarm.Activations != 2 || !slices.Equal(alarm.ActivationTimesNS, []int64{2000, 4000}) {
		t.Fatalf("report alarm = %+v; want two activations at 2000 and 4000", alarm)
	}
	raw, err := JSON(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	truth := decoded["truth"].(map[string]any)
	kpis := truth["kpis"].(map[string]any)
	entity := kpis["entities"].([]any)[0].(map[string]any)
	activation := entity["alarms"].([]any)[0].(map[string]any)
	if times := activation["activation_times_ns"].([]any); len(times) != 2 || times[0] != float64(2000) || times[1] != float64(4000) {
		t.Fatalf("serialized activation times = %v; want [2000 4000]", times)
	}
}

func TestAnOpenEvaluateReportDoesNotExposeAlarmActivationInstants(t *testing.T) {
	t.Parallel()
	report, err := Assemble(Folder{Observer: doc(openedEvaluate), Director: doc(activationTruth)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Truth != nil || report.TruthUnavailable != TruthSealed {
		t.Fatalf("open Evaluate report = %+v; want no truth and the sealed reason", report)
	}
}
