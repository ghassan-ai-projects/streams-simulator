package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestKPITruthRendersAsKPIAlarmAndAlarmSystemTables(t *testing.T) {
	t.Parallel()
	if sections := kpiSections(nil, nil); sections != nil {
		t.Fatalf("a run without KPI truth renders %v", sections)
	}
	truth := &model.KPITruth{Entities: []model.EntityKPIs{{EntityID: "CH-1",
		KPIs:   []model.KPIValue{{State: "cop", Mean: 4.25, Min: 3.5, Max: 5, Samples: 10}},
		Alarms: []model.AlarmKPI{{State: "temp_alarm", Priority: "high", Activations: 3, Chattering: true, ActiveS: 30, LongestActiveS: 20}}}},
		AlarmSystem: model.AlarmSystemKPIs{Activations: 3, PeakPer10Min: 3, Chattering: 1}}
	sections := kpiSections(&Truth{KPIs: truth}, nil)
	if len(sections) != 3 || sections[0].title != "KPIs" || sections[1].title != "Alarms" || sections[2].title != "Alarm system" {
		t.Fatalf("sections = %+v", sections)
	}
	for i, want := range []string{"| CH-1 | cop |  | 4.250 | 3.500 | 5.000 | 10 |", "| CH-1 | temp\\_alarm | high | 3 | yes | 30.000 | 20.000 | no |", "| 3 | 3 | 0 | 1 | 0 | 0 |"} {
		if got := strings.Join(sections[i].lines, "\n"); !strings.Contains(got, want) {
			t.Errorf("section %s:\n%s\nwant a row %s", sections[i].title, got, want)
		}
	}
}

func TestAKPIWhoseStartAndEndTheOutcomeRecordsShowsItsChangeWithItsUnit(t *testing.T) {
	t.Parallel()
	kpis := &model.KPITruth{Entities: []model.EntityKPIs{{EntityID: "B-1", KPIs: []model.KPIValue{
		{State: "delivered", Mean: 100067.5, Min: 100000, Max: 100104.473, Samples: 181}, {State: "load", Mean: 2, Min: 1, Max: 3, Samples: 181}}}}}
	scoring := &OutcomeScoring{Status: Scored, Result: &model.OutcomeReport{OutcomeScore: model.OutcomeScore{Runs: []model.OutcomeRun{
		{Mode: model.OutcomeDoNothing, Facts: &model.OutcomeFacts{Entities: []model.EntityOutcome{{EntityID: "B-1",
			KPIs: []model.KPIOutcome{{State: "delivered", Start: 100000, End: 100050}}}}}},
		{Mode: model.OutcomeAsRun, Facts: &model.OutcomeFacts{Entities: []model.EntityOutcome{{EntityID: "B-1",
			KPIs: []model.KPIOutcome{{State: "delivered", Start: 100000, End: 100104.4727}}}}}}}}}}
	lines := kpiSections(&Truth{KPIs: kpis, KPIUnits: map[string]string{"delivered": "t", "load": "kg"}}, asRunKPIs(scoring))[0].lines
	want := []string{"| Entity | KPI | Unit | Start | End | Change |", "|---|---|---|---|---|---|", "| B-1 | delivered | t | 100000.000 | 100104.473 | +104.473 |", "",
		"| Entity | KPI | Unit | Mean | Min | Max | Samples |", "|---|---|---|---|---|---|---|", "| B-1 | load | kg | 2.000 | 1.000 | 3.000 | 181 |"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("KPI section:\n%s\nwant\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
	if asRunKPIs(nil) != nil || asRunKPIs(&OutcomeScoring{Status: Failed}) != nil {
		t.Fatal("a report without a scored outcome has KPI ends")
	}
}

func TestAKPIUnitComesFromTheKindsThatDeclareItAndIsLeftOutWhenTheyDisagree(t *testing.T) {
	t.Parallel()
	artifact := &model.RunArtifact{DomainSpec: json.RawMessage(`{"kinds":[` +
		`{"name":"belt","kpis":["delivered","energy","load"],"state":[{"name":"delivered","unit":"t"},{"name":"energy","unit":"kWh"},{"name":"load"},{"name":"speed","unit":"m/s"}]},` +
		`{"name":"drive","kpis":["energy"],"state":[{"name":"energy","unit":"Wh"}]}]}`)}
	units := kpiUnits(&model.KPITruth{}, artifact)
	if len(units) != 1 || units["delivered"] != "t" {
		t.Fatalf("units = %v, want delivered in t only: energy has two units, load none and speed is no KPI", units)
	}
	if kpiUnits(nil, artifact) != nil || kpiUnits(&model.KPITruth{}, nil) != nil {
		t.Fatal("a run without KPI truth or without an embedded domain has KPI units")
	}
}

func TestTheReportReadsTheSealedKPITruthAndRefusesAMalformedOne(t *testing.T) {
	t.Parallel()
	sealed := journal.Record{Seq: 4, Type: journal.TypeKPITruth, Fields: map[string]any{"kpis": map[string]any{"to_ns": 9, "entities": []any{}}}}
	truth, err := truthOf(nil, []journal.Record{sealed})
	if err != nil || truth.KPIs == nil || truth.KPIs.ToNS != 9 {
		t.Fatalf("truth = %+v (%v), want the sealed KPI truth", truth, err)
	}
	sealed.Fields["kpis"] = "not an object"
	if _, err := truthOf(nil, []journal.Record{sealed}); err == nil || !strings.Contains(err.Error(), "record 4") {
		t.Fatalf("a malformed kpi.truth record gives %v, want an error naming the record", err)
	}
}
