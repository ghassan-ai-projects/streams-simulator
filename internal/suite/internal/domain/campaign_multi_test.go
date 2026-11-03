package domain

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func intPtr(n int) *int { return &n }

func multiSpec(phase CampaignPhase) *CampaignSpec {
	spec := smallSpec()
	spec.Phases = []CampaignPhase{phase}
	return spec
}

func threeFaults() map[string]float64 {
	return map[string]float64{"bearing_wear": 1, "misalignment": 1, "sensor_drift": 1}
}

func buildPlans(t *testing.T, spec *CampaignSpec) []model.SessionPlan {
	t.Helper()
	_, plans, err := BuildCampaign(spec, pumpStation(t))
	if err != nil {
		t.Fatal(err)
	}
	return plans
}

func TestARunCarriesItsFaultsAndNegativesOnDistinctEntitiesInsideTheWindow(t *testing.T) {
	t.Parallel()
	phase := CampaignPhase{Name: "many", Domain: "pump-station", Runs: 6, Faults: threeFaults(), FaultsPerRun: intPtr(2), NegativesPerRun: 1}
	plans := buildPlans(t, multiSpec(phase))
	positives := map[string]int{}
	for i, plan := range plans {
		if len(plan.Disturbances) != 3 {
			t.Fatalf("run %d holds %d disturbances, want 2 faults and 1 negative", i+1, len(plan.Disturbances))
		}
		instances := map[string]bool{}
		negatives := 0
		for _, d := range plan.Disturbances {
			instances[d.Instance] = true
			negatives += countIf(d.Fault == "transient_none")
			if d.Fault != "transient_none" {
				positives[d.Fault]++
			}
			if d.AtS < 0 || d.AtS >= 43200 || (d.Fault == "transient_none") != (d.DurationS != nil) {
				t.Fatalf("run %d disturbance %+v", i+1, d)
			}
		}
		if len(instances) != 3 || negatives != 1 {
			t.Fatalf("run %d: %d distinct entities and %d negatives, want 3 and 1: %+v", i+1, len(instances), negatives, plan.Disturbances)
		}
	}
	if want := map[string]int{"bearing_wear": 4, "misalignment": 4, "sensor_drift": 4}; !maps.Equal(positives, want) {
		t.Fatalf("fault slots = %v, want %v", positives, want)
	}
}

func countIf(condition bool) int {
	if condition {
		return 1
	}
	return 0
}

func TestSeveralFaultsPerRunAreDeterministicFromTheSeed(t *testing.T) {
	t.Parallel()
	phase := CampaignPhase{Name: "many", Domain: "pump-station", Runs: 8, Faults: threeFaults(), FaultsPerRun: intPtr(3)}
	first, second := buildPlans(t, multiSpec(phase)), buildPlans(t, multiSpec(phase))
	if !slices.EqualFunc(first, second, func(a, b model.SessionPlan) bool {
		return slices.EqualFunc(a.Disturbances, b.Disturbances, func(x, y model.Disturbance) bool {
			return x.Instance == y.Instance && x.Fault == y.Fault && x.AtS == y.AtS
		})
	}) {
		t.Fatal("two builds of one spec differ")
	}
}

func TestAWarmUpPhaseHoldsNoFaultsAndAtMostNegatives(t *testing.T) {
	t.Parallel()
	warmup := CampaignPhase{Name: "warmup", Domain: "pump-station", Runs: 2, FaultsPerRun: intPtr(0)}
	for _, plan := range buildPlans(t, multiSpec(warmup)) {
		if len(plan.Disturbances) != 0 {
			t.Fatalf("a warm-up run holds %+v", plan.Disturbances)
		}
	}
	quiet := CampaignPhase{Name: "quiet", Domain: "pump-station", Runs: 3, FaultsPerRun: intPtr(0), NegativesPerRun: 2}
	for _, plan := range buildPlans(t, multiSpec(quiet)) {
		if len(plan.Disturbances) != 2 || plan.Disturbances[0].Fault != "transient_none" || plan.Disturbances[1].Fault != "transient_none" {
			t.Fatalf("a negatives-only run holds %+v", plan.Disturbances)
		}
	}
}

func TestAPhaseThatCrowdsOrMisusesThePerRunCountsIsRefused(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		change func(*CampaignPhase)
		reason string
	}{
		{"more disturbances than entities", func(p *CampaignPhase) { p.FaultsPerRun = intPtr(40) }, "distinct entities"},
		{"negatives push past the entities", func(p *CampaignPhase) { p.NegativesPerRun = 40 }, "distinct entities"},
		{"a share beside several faults", func(p *CampaignPhase) { p.FaultsPerRun, p.NegativeShare = intPtr(2), 0.5 }, "negative_share needs faults_per_run 1"},
		{"a negative count", func(p *CampaignPhase) { p.NegativesPerRun = -1 }, "faults_per_run and negatives_per_run"},
		{"a count beyond a plan", func(p *CampaignPhase) { p.FaultsPerRun = intPtr(1 << 62) }, "faults_per_run and negatives_per_run"},
		{"weights without faults", func(p *CampaignPhase) { p.FaultsPerRun = intPtr(0) }, "faults_per_run is 0"},
		{"no weights for the faults", func(p *CampaignPhase) { p.FaultsPerRun, p.Faults = intPtr(2), nil }, "lists no fault"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			phase := CampaignPhase{Name: "p", Domain: "pump-station", Runs: 2, Faults: threeFaults()}
			tc.change(&phase)
			var invalid *CampaignSpecError
			if _, _, err := BuildCampaign(multiSpec(phase), pumpStation(t)); !errors.As(err, &invalid) || !strings.Contains(invalid.Message, tc.reason) {
				t.Fatalf("BuildCampaign = %v, want a CampaignSpecError saying %q", err, tc.reason)
			}
		})
	}
}
