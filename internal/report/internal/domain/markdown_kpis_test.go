package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestKPITruthRendersAsKPIAlarmAndAlarmSystemTables(t *testing.T) {
	t.Parallel()
	if sections := kpiSections(nil); sections != nil {
		t.Fatalf("a run without KPI truth renders %v", sections)
	}
	truth := &model.KPITruth{Entities: []model.EntityKPIs{{EntityID: "CH-1",
		KPIs:   []model.KPIValue{{State: "cop", Mean: 4.25, Min: 3.5, Max: 5, Samples: 10}},
		Alarms: []model.AlarmKPI{{State: "temp_alarm", Priority: "high", Activations: 3, Chattering: true, ActiveS: 30, LongestActiveS: 20}}}},
		AlarmSystem: model.AlarmSystemKPIs{Activations: 3, PeakPer10Min: 3, Chattering: 1}}
	sections := kpiSections(truth)
	if len(sections) != 3 || sections[0].title != "KPIs" || sections[1].title != "Alarms" || sections[2].title != "Alarm system" {
		t.Fatalf("sections = %+v", sections)
	}
	for i, want := range []string{"| CH-1 | cop | 4.250 | 3.500 | 5.000 | 10 |", "| CH-1 | temp\\_alarm | high | 3 | yes | 30.000 | 20.000 | no |", "| 3 | 3 | 0 | 1 | 0 | 0 |"} {
		if got := strings.Join(sections[i].lines, "\n"); !strings.Contains(got, want) {
			t.Errorf("section %s:\n%s\nwant a row %s", sections[i].title, got, want)
		}
	}
}
