package domain

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpStation(t *testing.T) map[string]*domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.PumpStationDomain())
	if err != nil {
		t.Fatal(err)
	}
	return map[string]*domain.Compiled{"pump-station": spec}
}

func smallSpec() *CampaignSpec {
	faults := map[string]float64{"bearing_wear": 1, "misalignment": 1, "sensor_drift": 1}
	return &CampaignSpec{SchemaVersion: CampaignSpecVersion, Seed: 11, HostSocket: "/tmp/h.sock", DurationS: 172800, FaultWindowS: 43200,
		NegativeDurationS: 3600, Phases: []CampaignPhase{
			{Name: "learn", Domain: "pump-station", Runs: 10, NegativeShare: 0.4, Faults: faults},
			{Name: "holdout", Domain: "pump-station", Runs: 5, NegativeShare: 0.4, Faults: faults},
		}}
}

func faultCounts(campaign *Campaign, plans []model.SessionPlan, phase string) map[string]int {
	counts := map[string]int{}
	for i, run := range campaign.Runs {
		if run.Phase == phase {
			counts[plans[i].Disturbances[0].Fault]++
		}
	}
	return counts
}

func TestApportionSplitsRunsByWeightWithLargestRemainders(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		weights map[string]float64
		total   int
		want    []string
	}{
		{"equal and divisible", map[string]float64{"a": 1, "b": 1}, 4, []string{"a", "a", "b", "b"}},
		{"leftover goes to the first of equal remainders", map[string]float64{"a": 1, "b": 1, "c": 1}, 4, []string{"a", "a", "b", "c"}},
		{"weights decide", map[string]float64{"a": 3, "b": 1}, 4, []string{"a", "a", "a", "b"}},
		{"biggest remainder wins the leftover", map[string]float64{"a": 5, "b": 3}, 3, []string{"a", "a", "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := apportion(tc.weights, tc.total); !slices.Equal(got, tc.want) {
				t.Fatalf("apportion = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestACampaignHoldsExactCountsPerPhaseInOneOrder(t *testing.T) {
	t.Parallel()
	campaign, plans, err := BuildCampaign(smallSpec(), pumpStation(t))
	if err != nil || len(campaign.Runs) != 15 || len(plans) != 15 {
		t.Fatalf("BuildCampaign = %d runs, %d plans, %v", len(campaign.Runs), len(plans), err)
	}
	learn := map[string]int{"bearing_wear": 2, "misalignment": 2, "sensor_drift": 2, "transient_none": 4}
	holdout := map[string]int{"bearing_wear": 1, "misalignment": 1, "sensor_drift": 1, "transient_none": 2}
	if got := faultCounts(campaign, plans, "learn"); !reflect.DeepEqual(got, learn) {
		t.Fatalf("learn faults = %v, want %v", got, learn)
	}
	if got := faultCounts(campaign, plans, "holdout"); !reflect.DeepEqual(got, holdout) {
		t.Fatalf("holdout faults = %v, want %v", got, holdout)
	}
	for i, run := range campaign.Runs {
		if run.Index != i+1 || run.SessionPlan != PlanFileName(i+1) || (i < 10) != (run.Phase == "learn") {
			t.Fatalf("run %d = %+v, want index %d and its phase in order", i, run, i+1)
		}
	}
}

func TestACampaignIsAPureFunctionOfItsSpec(t *testing.T) {
	t.Parallel()
	first, firstPlans, err := BuildCampaign(smallSpec(), pumpStation(t))
	second, secondPlans, again := BuildCampaign(smallSpec(), pumpStation(t))
	if err != nil || again != nil || !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstPlans, secondPlans) {
		t.Fatalf("two builds of one spec differ: %v %v", err, again)
	}
	other := smallSpec()
	other.Seed++
	changed, _, err := BuildCampaign(other, pumpStation(t))
	if err != nil || reflect.DeepEqual(first.Runs, changed.Runs) {
		t.Fatalf("another seed gave the same runs (%v)", err)
	}
}

func TestOnlyTheTransientEndsOnItsOwnAndEveryDisturbanceStartsInTheWindow(t *testing.T) {
	t.Parallel()
	campaign, plans, err := BuildCampaign(smallSpec(), pumpStation(t))
	if err != nil {
		t.Fatal(err)
	}
	for i, run := range campaign.Runs {
		d := plans[i].Disturbances[0]
		if d.AtS < 0 || d.AtS >= 43200 || (d.Fault == "transient_none") != (d.DurationS != nil) {
			t.Fatalf("run %d disturbance %+v", run.Index, d)
		}
		if plans[i].Seed == nil || *plans[i].Seed != run.Seed || plans[i].Domain != run.Domain || len(plans[i].Disturbances) != 1 {
			t.Fatalf("run %d plan %+v does not carry the run", run.Index, plans[i])
		}
	}
}

func TestAScheduleDrawsFromTheSeedAlone(t *testing.T) {
	t.Parallel()
	phase := &smallSpec().Phases[0]
	first := phase.schedule("transient_none", randutil.NewSplitMix64(1))
	second := phase.schedule("transient_none", randutil.NewSplitMix64(1))
	if !slices.Equal(first, second) || len(first) != 10 {
		t.Fatalf("schedules = %v and %v", first, second)
	}
}

func TestASpecThatBreaksARuleIsRefusedWithItsReason(t *testing.T) {
	t.Parallel()
	for name, change := range map[string]func(*CampaignSpec){
		"another version":       func(s *CampaignSpec) { s.SchemaVersion = "0.2" },
		"no host socket":        func(s *CampaignSpec) { s.HostSocket = "" },
		"window past the run":   func(s *CampaignSpec) { s.FaultWindowS = s.DurationS },
		"no phases":             func(s *CampaignSpec) { s.Phases = nil },
		"a phase twice":         func(s *CampaignSpec) { s.Phases[1].Name = "learn" },
		"no runs":               func(s *CampaignSpec) { s.Phases[0].Runs = 0 },
		"share above one":       func(s *CampaignSpec) { s.Phases[0].NegativeShare = 1.5 },
		"unknown domain":        func(s *CampaignSpec) { s.Phases[0].Domain = "no-such-domain" },
		"undeclared fault":      func(s *CampaignSpec) { s.Phases[0].Faults["no_such_fault"] = 1 },
		"the negative as a hit": func(s *CampaignSpec) { s.Phases[0].Faults["transient_none"] = 1 },
		"a commanded state":     func(s *CampaignSpec) { s.Phases[0].Faults["overload"] = 1 },
		"zero weight":           func(s *CampaignSpec) { s.Phases[0].Faults["bearing_wear"] = 0 },
		"no positive fault":     func(s *CampaignSpec) { s.Phases[0].Faults = nil },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			spec := smallSpec()
			spec.Phases[1].Faults = maps.Clone(spec.Phases[1].Faults)
			spec.Phases[0].Faults = maps.Clone(spec.Phases[0].Faults)
			change(spec)
			var invalid *CampaignSpecError
			if _, _, err := BuildCampaign(spec, pumpStation(t)); !errors.As(err, &invalid) {
				t.Fatalf("BuildCampaign = %v, want a CampaignSpecError", err)
			}
		})
	}
}

func TestAPulseNegativeFaultKeepsItsOwnDurationInThePlan(t *testing.T) {
	t.Parallel()
	domains := pumpStation(t)
	kind := domains["pump-station"].Spec.Kinds[0]
	for i := range kind.Faults {
		if kind.Faults[i].ID == "transient_none" {
			kind.Faults[i].Onset.Shape = model.ShapePulse
		}
	}
	_, plans, err := BuildCampaign(smallSpec(), domains)
	if err != nil {
		t.Fatal(err)
	}
	for _, plan := range plans {
		if d := plan.Disturbances[0]; d.Fault == "transient_none" && d.DurationS != nil {
			t.Fatalf("plan %+v gives a pulse fault a duration, which plantwire connect refuses", plan)
		}
	}
}
