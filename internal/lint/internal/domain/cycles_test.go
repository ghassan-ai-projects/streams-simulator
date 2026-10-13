package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func stagedKind(pressure *model.F0Dyn) *model.Kind {
	return &model.Kind{Name: "compressor", Dynamics: []model.Dynamics{
		{Target: "pressure", Tier: "F0", F0: pressure},
		{Target: "loaded", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormSequencer, Stages: 1, HoldS: 10,
			StageUp:   &model.Condition{State: "pressure", Operator: "lt", Threshold: 1.5},
			StageDown: &model.Condition{State: "pressure", Operator: "gt", Threshold: 2.5}}},
		{Target: "flow", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: "first_order_lag", Gain: 1, TimeConstantS: 12,
			Inputs: []model.F1Input{{State: "loaded", Coef: 700}}}},
	}}
}

func TestAStateStagedByASequencerOnACyclingConditionCyclesAndIsNotExpectedToSettle(t *testing.T) {
	t.Parallel()
	spec := &domain.Compiled{Spec: &model.DomainSpec{}}
	if !cyclic(spec, stagedKind(dailyCycle()), "flow") {
		t.Fatal("a flow staged on a daily pressure cycle is not found cyclic, so the lint would expect it to settle")
	}
	if cyclic(spec, stagedKind(&model.F0Dyn{Baseline: 2}), "flow") {
		t.Fatal("a flow staged on a constant pressure is found cyclic, so the lint would never check that it settles")
	}
}
