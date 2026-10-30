package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func wearLabel() *model.GroundTruthRecord {
	return &model.GroundTruthRecord{EntityID: "site-a/pump-1", Label: "bearing_wear", ExpectedEffector: "schedule_maintenance"}
}

func maintenance(workOrder any) world.EffectorCall {
	return world.EffectorCall{Effector: "schedule_maintenance", EntityID: "site-a/pump-1", EffectApplied: true,
		Args: map[string]any{"pump_id": "site-a/pump-1", "work_order": workOrder}}
}

func TestAMaintenanceCallIsAppropriateOnlyWhenItsWorkOrderNamesTheFault(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.PumpStationDomain())
	if err != nil {
		t.Fatal(err)
	}
	ev := Evidence{Domain: spec}
	matchArg := responderArg(ev, wearLabel())
	if matchArg != "work_order" {
		t.Fatalf("responder argument = %q, want work_order", matchArg)
	}
	for _, tc := range []struct {
		name string
		call world.EffectorCall
		want bool
	}{
		{"the fault named", maintenance("bearing_wear"), true},
		{"another fault named", maintenance("misalignment"), false},
		{"no work order", maintenance(nil), false},
		{"a number", maintenance(3.0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := loopFrom(&model.Verdict{}, wearLabel(), []world.EffectorCall{tc.call}, matchArg)
			if got.ActionAppropriate != tc.want || got.EffectCalls != 1 {
				t.Fatalf("loop = %+v, want appropriate %v", got, tc.want)
			}
			if right := rightCall(tc.call, wearLabel(), matchArg); right != tc.want {
				t.Fatalf("rightCall = %v, want %v", right, tc.want)
			}
		})
	}
}

func TestAnEffectorWithoutAResponderIsJudgedOnEffectorAndEntityAlone(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.PumpStationDomain())
	if err != nil {
		t.Fatal(err)
	}
	gt := &model.GroundTruthRecord{EntityID: "site-a/pump-1", Label: "overload", ExpectedEffector: "reduce_load"}
	if arg := responderArg(Evidence{Domain: spec}, gt); arg != "" {
		t.Fatalf("responder argument of reduce_load = %q, want none", arg)
	}
	call := world.EffectorCall{Effector: "reduce_load", EntityID: "site-a/pump-1", EffectApplied: true}
	if !rightCall(call, gt, "") {
		t.Fatal("reduce_load on the right pump was not credited")
	}
}
