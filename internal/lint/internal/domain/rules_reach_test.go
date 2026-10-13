package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func dailyCycle() *model.F0Dyn {
	return &model.F0Dyn{Baseline: 2, Seasonality: []model.Seasonality{{PeriodS: 86400, Amplitude: 1}}}
}

func followingASeasonalFeeder(spec *model.DomainSpec) {
	plant := plantKind(spec)
	plant.Dynamics = append(plant.Dynamics, model.Dynamics{Target: "u", Tier: "F0", F0: &model.F0Dyn{Baseline: 2},
		Follow: &model.Follow{Link: "plant__u__feeder", State: "u", Aggregate: model.AggregateSum}})
	spec.Kinds = append(spec.Kinds, feederKind(plant))
	spec.Instances = append(spec.Instances, model.InstanceSet{Kind: "feeder", IDTemplate: "f-{n}", Count: model.EntityCount{Default: 1}})
	spec.LinkTypes = []model.LinkType{{Name: "plant__u__feeder", From: "feeder", To: "plant"}}
	spec.Links = []model.Link{{Type: "plant__u__feeder", From: "f-1", To: "p-1"}, {Type: "plant__u__feeder", From: "f-1", To: "p-2"}}
	lagOf(spec).TimeConstantS = 3600
}

func feederKind(plant *model.Kind) model.Kind {
	feeder := *plant
	feeder.Name = "feeder"
	feeder.Dynamics = []model.Dynamics{{Target: "u", Tier: "F0", F0: dailyCycle()}}
	feeder.Channels = append([]model.Channel{}, plant.Channels...)
	feeder.Channels[0].Name, feeder.Channels[1].Name = "feeder_in", "feeder_out"
	feeder.Faults = append([]model.Fault{}, plant.Faults...)
	feeder.Faults[0].ID = "feeder_leak"
	feeder.Faults[0].Observability.Detector.Inputs, feeder.Faults[0].Observability.Detector.Outputs = []string{"feeder_in"}, []string{"feeder_out"}
	return feeder
}

func exchangingWithASeasonallyFedStore(spec *model.DomainSpec) {
	plant := plantKind(spec)
	plant.State = append(plant.State, model.State{Name: "a", Min: new(float64)}, model.State{Name: "b", Min: new(float64)})
	plant.Dynamics = append(plant.Dynamics,
		model.Dynamics{Target: "u", Tier: "F0", F0: dailyCycle()},
		model.Dynamics{Target: "a", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormStore, Capacity: 100, Inputs: []model.F1Input{{State: "u", Coef: 1}}}},
		model.Dynamics{Target: "b", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormStore, Capacity: 100}})
	plant.Flows = []model.Flow{{Name: "ab", From: "a", To: "b", Conductance: 0.01}}
}

func TestAStateFollowingASeasonalProviderCyclesAndIsNotExpectedToSettle(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 24, followingASeasonalFeeder)
	if finding, drifting := findingFor(report, codeNotSettled, "p-1", "x"); drifting {
		t.Fatalf("a state driven through a follow of a daily cycle was reported as unsettled: %+v", finding)
	}
}

func TestAStateFollowingAConstantProviderIsStillExpectedToSettle(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 2, func(spec *model.DomainSpec) {
		followingASeasonalFeeder(spec)
		spec.Kinds[1].Dynamics = []model.Dynamics{{Target: "u", Tier: "F0", F0: &model.F0Dyn{Baseline: 2}}}
		plantKind(spec).State[1].Initial = 0
		lagOf(spec).TimeConstantS = 3_600_000
	})
	if _, drifting := findingFor(report, codeNotSettled, "p-1", "x"); !drifting {
		t.Fatalf("a state behind a constant provider that still ramps was not reported: %+v", report.Findings)
	}
}

func TestAStoreFedThroughAFlowFromASeasonalStoreCyclesAndIsNotExpectedToSettle(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 12, exchangingWithASeasonallyFedStore)
	for _, state := range []string{"a", "b"} {
		if finding, drifting := findingFor(report, codeNotSettled, "p-1", state); drifting {
			t.Errorf("store %s, driven by a daily cycle directly or through a flow, was reported as unsettled: %+v", state, finding)
		}
	}
}
