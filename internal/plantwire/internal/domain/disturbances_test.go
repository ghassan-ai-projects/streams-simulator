package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func disturbedInstances() []PlanInstance {
	instances := benchInstances()
	instances[0].Faults = []PlanFault{{ID: "zone_heat_load"}, {ID: "dht11_stuck"}}
	instances[1].Faults = []PlanFault{{ID: "led_flicker", Pulse: true}}
	return instances
}

func seconds(value float64) *float64 { return &value }

func TestEveryBrokenDisturbanceIsReportedAtItsMemberWithItsRule(t *testing.T) {
	t.Parallel()
	zone := model.RealBinding{Instance: "zone-01", Board: "uno-07", Device: "zone-01"}
	cases := map[string]struct {
		real             []model.RealBinding
		disturbance      model.Disturbance
		pointer, keyword string
	}{
		"an unknown instance": {nil, model.Disturbance{AtS: 5, Instance: "ghost", Fault: "zone_heat_load"}, "/disturbances/1/instance", "unknown_instance"},
		"a fault the kind does not declare": {nil, model.Disturbance{AtS: 5, Instance: "led-01", Fault: "zone_heat_load"},
			"/disturbances/1/fault", KeywordUnknownFault},
		"an instance bound to a real board": {[]model.RealBinding{zone}, model.Disturbance{AtS: 5, Instance: "zone-01", Fault: "zone_heat_load"},
			"/disturbances/1/instance", KeywordDisturbanceOnRealInstance},
		"a pulse fault with a duration": {nil, model.Disturbance{AtS: 5, Instance: "led-01", Fault: "led_flicker", DurationS: seconds(3)},
			"/disturbances/1/duration_s", KeywordDurationWithPulseFault},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan := benchPlan(c.real...)
			plan.Disturbances = []model.Disturbance{{AtS: 1, Instance: "zone-01", Fault: "dht11_stuck"}, c.disturbance}
			if len(c.real) > 0 {
				plan.Disturbances[0].Instance = "led-01"
				plan.Disturbances[0].Fault = "led_flicker"
			}
			_, issues := CheckPlan(plan, disturbedInstances())
			if len(issues) != 1 || issues[0].Pointer != c.pointer || issues[0].Keyword != c.keyword {
				t.Fatalf("issues %+v, want one %s at %s", issues, c.keyword, c.pointer)
			}
		})
	}
}

func TestAPlanOfSimulatedDisturbancesHoldsAndIsScheduledFromTheWorldStart(t *testing.T) {
	t.Parallel()
	plan := benchPlan()
	plan.Disturbances = []model.Disturbance{
		{AtS: 120, Instance: "zone-01", Fault: "zone_heat_load", DurationS: seconds(600), Severity: seconds(0.5)},
		{AtS: 0.0000000014, Instance: "led-01", Fault: "led_flicker"},
		{AtS: 120, Instance: "zone-01", Fault: "dht11_stuck"},
	}
	if _, issues := CheckPlan(plan, disturbedInstances()); len(issues) != 0 {
		t.Fatalf("issues %+v, want none", issues)
	}
	const start = 1_767_225_600_000_000_000
	want := []ScheduledFault{
		{Entity: "zone-01", Fault: "zone_heat_load", OnsetNS: start + 120e9, ClearNS: start + 720e9, Params: map[string]any{"severity": 0.5}},
		{Entity: "led-01", Fault: "led_flicker", OnsetNS: start + 1},
		{Entity: "zone-01", Fault: "dht11_stuck", OnsetNS: start + 120e9},
	}
	if got := ScheduleDisturbances(plan.Disturbances, start); !reflect.DeepEqual(got, want) {
		t.Fatalf("schedule %+v, want %+v in the plan's order", got, want)
	}
	if got := ScheduleDisturbances(nil, start); len(got) != 0 {
		t.Fatalf("no disturbances schedule %+v", got)
	}
}
