package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const sensorDigest = "sha256:7d0921b1526ddbb40760f3244b7b5f0a6443a125da593a9f0fa2878a871d9d48"

func benchInstances() []PlanInstance {
	return []PlanInstance{{ID: "zone-01", Kind: "thermal_sensor", Implements: "thermal_sensor", KindDigest: sensorDigest},
		{ID: "led-01", Kind: "indicator", Implements: "indicator", KindDigest: "sha256:led"},
		{ID: "lamp-01", Kind: "lamp"}}
}

func benchPlan(bindings ...model.RealBinding) *model.SessionPlan {
	return &model.SessionPlan{SchemaVersion: "0.1", Domain: "arduino-thermal", HostSocket: "/tmp/h.sock", GatewaySocket: "/tmp/g.sock", Real: bindings}
}

func TestAPlanBindsARealBoardToAnInstanceWhoseInterfaceIsAKindFile(t *testing.T) {
	t.Parallel()
	plan, issues := CheckPlan(benchPlan(model.RealBinding{Instance: "zone-01", Board: "uno-07", Device: "zone-01"}), benchInstances())
	if len(issues) != 0 || plan.Environment != EnvironmentMixed {
		t.Fatalf("issues %+v, environment %q, want a clean mixed plan", issues, plan.Environment)
	}
	want := []PlannedInstance{
		{Instance: "zone-01", Origin: PlannedReal, Kind: "thermal_sensor", KindDigest: sensorDigest, Board: "uno-07", Device: "zone-01"},
		{Instance: "led-01", Origin: PlannedSimulated, Kind: "indicator"},
		{Instance: "lamp-01", Origin: PlannedSimulated, Kind: "lamp"},
	}
	if !reflect.DeepEqual(plan.Instances, want) {
		t.Fatalf("instances = %+v, want %+v", plan.Instances, want)
	}
	bound := []Bound{{Index: 0, Instance: "zone-01", Board: "uno-07", Device: "zone-01", Kind: "thermal_sensor", KindDigest: sensorDigest}}
	if !reflect.DeepEqual(plan.Bound, bound) {
		t.Fatalf("bound = %+v, want %+v", plan.Bound, bound)
	}
}

func TestAPlanWithNoGatewayIsAllSimulated(t *testing.T) {
	t.Parallel()
	plan, issues := CheckPlan(&model.SessionPlan{Domain: "arduino-thermal", HostSocket: "/tmp/h.sock"}, benchInstances())
	if len(issues) != 0 || plan.Environment != EnvironmentSimulated || len(plan.Bound) != 0 || plan.Instances[0].Origin != PlannedSimulated {
		t.Fatalf("plan %+v, issues %+v, want every instance simulated", plan, issues)
	}
}

func TestEveryBrokenBindingIsReportedAtItsMemberWithItsRule(t *testing.T) {
	t.Parallel()
	zone := model.RealBinding{Instance: "zone-01", Board: "uno-07", Device: "zone-01"}
	cases := map[string]struct {
		plan             *model.SessionPlan
		pointer, keyword string
	}{
		"an unknown instance": {benchPlan(model.RealBinding{Instance: "ghost", Board: "uno-07", Device: "d"}), "/real/0/instance", "unknown_instance"},
		"no kind file":        {benchPlan(model.RealBinding{Instance: "lamp-01", Board: "uno-07", Device: "d"}), "/real/0/instance", "no_kind_file"},
		"an instance twice":   {benchPlan(zone, model.RealBinding{Instance: "zone-01", Board: "uno-08", Device: "z"}), "/real/1/instance", "instance_bound_twice"},
		"a device twice":      {benchPlan(zone, model.RealBinding{Instance: "led-01", Board: "uno-07", Device: "zone-01"}), "/real/1/device", "device_bound_twice"},
		"no gateway":          {&model.SessionPlan{HostSocket: "/tmp/h.sock", Real: []model.RealBinding{zone}}, "/real", "gateway_required"},
		"one socket for both": {&model.SessionPlan{HostSocket: "/tmp/h.sock", GatewaySocket: "/tmp/h.sock"}, "/gateway_socket", "same_socket"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan, issues := CheckPlan(tc.plan, benchInstances())
			if len(issues) != 1 || issues[0].Pointer != tc.pointer || issues[0].Keyword != tc.keyword || issues[0].Message == "" {
				t.Fatalf("issues = %+v, want one %s at %s", issues, tc.keyword, tc.pointer)
			}
			for _, b := range plan.Bound {
				if b.Index != 0 {
					t.Fatalf("bound = %+v, want a broken binding never bound", plan.Bound)
				}
			}
		})
	}
}
