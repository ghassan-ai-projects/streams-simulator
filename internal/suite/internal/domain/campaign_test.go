package domain

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
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
	for _, tc := range []struct {
		name   string
		change func(*CampaignSpec)
		reason string
	}{
		{"another version", func(s *CampaignSpec) { s.SchemaVersion = "0.2" }, "schema_version"},
		{"a seed beyond 2^53", func(s *CampaignSpec) { s.Seed = 1 << 60 }, "seed must be at most"},
		{"no host socket", func(s *CampaignSpec) { s.HostSocket = "" }, "host_socket"},
		{"window past the run", func(s *CampaignSpec) { s.FaultWindowS = s.DurationS }, "must fit inside duration_s"},
		{"no phases", func(s *CampaignSpec) { s.Phases = nil }, "at least one phase"},
		{"a phase twice", func(s *CampaignSpec) { s.Phases[1].Name = "learn" }, "listed twice"},
		{"no runs", func(s *CampaignSpec) { s.Phases[0].Runs = 0 }, "at least one run"},
		{"share above one", func(s *CampaignSpec) { s.Phases[0].NegativeShare = 1.5 }, "negative_share"},
		{"unknown domain", func(s *CampaignSpec) { s.Phases[0].Domain = "no-such-domain" }, "not available"},
		{"undeclared fault", func(s *CampaignSpec) { s.Phases[0].Faults["no_such_fault"] = 1 }, `"no_such_fault"`},
		{"the negative as a hit", func(s *CampaignSpec) { s.Phases[0].Faults["transient_none"] = 1 }, `"transient_none"`},
		{"a commanded state", func(s *CampaignSpec) { s.Phases[0].Faults["overload"] = 1 }, "D-116"},
		{"a weight beyond range", func(s *CampaignSpec) { s.Phases[0].Faults["bearing_wear"] = 1e308 }, "at most 1e+06"},
		{"a duration beyond 30 d", func(s *CampaignSpec) { s.DurationS = 1e9 }, "duration_s must be above 0"},
		{"a transient past run", func(s *CampaignSpec) { s.NegativeDurationS = s.DurationS }, "must fit inside"},
		{"a phase of billions", func(s *CampaignSpec) { s.Phases[0].Runs = 1_000_000_000_000 }, "at most 4096 runs"},
		{"too many runs", func(s *CampaignSpec) { s.Phases[0].Runs, s.Phases[1].Runs = 3000, 3000 }, "a campaign has at most 4096"},
		{"a long phase name", func(s *CampaignSpec) { s.Phases[0].Name = strings.Repeat("p", 65) }, "64 characters"},
		{"zero weight", func(s *CampaignSpec) { s.Phases[0].Faults["bearing_wear"] = 0 }, `"bearing_wear"`},
		{"no positive fault", func(s *CampaignSpec) { s.Phases[0].Faults = nil }, "lists no fault"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			spec := smallSpec()
			spec.Phases[1].Faults = maps.Clone(spec.Phases[1].Faults)
			spec.Phases[0].Faults = maps.Clone(spec.Phases[0].Faults)
			tc.change(spec)
			var invalid *CampaignSpecError
			if _, _, err := BuildCampaign(spec, pumpStation(t)); !errors.As(err, &invalid) || !strings.Contains(invalid.Message, tc.reason) {
				t.Fatalf("BuildCampaign = %v, want a CampaignSpecError saying %q", err, tc.reason)
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

func TestEveryNegativeClassFaultIsRefusedAsAPositiveOne(t *testing.T) {
	t.Parallel()
	domains := pumpStation(t)
	kind := domains["pump-station"].Spec.Kinds[0]
	for i := range kind.Faults {
		if kind.Faults[i].ID == "misalignment" {
			kind.Faults[i].IsNegativeClass = true
		}
	}
	spec := smallSpec()
	spec.Phases[0].Faults = map[string]float64{"bearing_wear": 1, "transient_none": 1}
	var invalid *CampaignSpecError
	if _, _, err := BuildCampaign(spec, domains); !errors.As(err, &invalid) || !strings.Contains(invalid.Message, `"transient_none"`) {
		t.Fatalf("BuildCampaign = %v, want transient_none refused as a positive fault", err)
	}
}

func TestNegativeRunsAreRefusedWhenNoInstanceCanCarryTheNegativeFault(t *testing.T) {
	t.Parallel()
	domains := pumpStation(t)
	spare := domains["pump-station"].Spec.Kinds[0]
	spare.Name = "spare"
	var carried, kept []model.Fault
	for _, fault := range spare.Faults {
		if fault.IsNegativeClass {
			carried = append(carried, fault)
		} else {
			kept = append(kept, fault)
		}
	}
	spare.Faults = carried
	domains["pump-station"].Spec.Kinds[0].Faults = kept
	domains["pump-station"].Spec.Kinds = append(domains["pump-station"].Spec.Kinds, spare)
	var invalid *CampaignSpecError
	if _, _, err := BuildCampaign(smallSpec(), domains); !errors.As(err, &invalid) || !strings.Contains(invalid.Message, "no instance") {
		t.Fatalf("BuildCampaign = %v, want a CampaignSpecError saying no instance can carry the fault", err)
	}
}
