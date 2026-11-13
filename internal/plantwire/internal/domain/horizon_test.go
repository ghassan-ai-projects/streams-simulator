package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func simulatedPlan(horizonS float64, disturbances ...model.Disturbance) *model.SessionPlan {
	plan := benchPlan()
	plan.GatewaySocket = ""
	plan.HorizonS = &horizonS
	plan.Disturbances = disturbances
	return plan
}

func TestASessionEndsAtItsHorizonOrRunsOpenEnded(t *testing.T) {
	t.Parallel()
	const start = 1_767_225_600_000_000_000
	if got := SessionEndNS(simulatedPlan(57600.5), start); got != start+57_600_500_000_000 {
		t.Fatalf("end = %d, want the world start plus the horizon", got)
	}
	if got := SessionEndNS(benchPlan(), start); got != start+openEndedNS || got-start < 100*365*24*3600*1e9 {
		t.Fatalf("end = %d, want a hundred years: a plan without a horizon runs until it is stopped", got)
	}
}

func TestAHorizonHoldsEveryDisturbanceAndOnlyASimulatedSession(t *testing.T) {
	t.Parallel()
	heat := model.Disturbance{AtS: 120, Instance: "zone-01", Fault: "zone_heat_load", DurationS: seconds(600)}
	late, lasting := heat, heat
	late.AtS = 900
	lasting.DurationS = seconds(781)
	gateway := simulatedPlan(900)
	gateway.GatewaySocket = "/tmp/g.sock"
	cases := map[string]struct {
		plan             *model.SessionPlan
		pointer, keyword string
	}{
		"a disturbance at the horizon":            {simulatedPlan(900, heat, late), "/disturbances/1/at_s", KeywordDisturbanceAfterHorizon},
		"a disturbance cleared after the horizon": {simulatedPlan(900, lasting), "/disturbances/0/duration_s", KeywordClearAfterHorizon},
		"a horizon on a session with a gateway":   {gateway, "/horizon_s", KeywordHorizonWithGateway},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, issues := CheckPlan(c.plan, disturbedInstances())
			if len(issues) != 1 || issues[0].Pointer != c.pointer || issues[0].Keyword != c.keyword {
				t.Fatalf("issues %+v, want one %s at %s", issues, c.keyword, c.pointer)
			}
		})
	}
	if _, issues := CheckPlan(simulatedPlan(720, heat), disturbedInstances()); len(issues) != 0 {
		t.Fatalf("issues %+v, want a disturbance cleared exactly at the horizon to hold", issues)
	}
}
