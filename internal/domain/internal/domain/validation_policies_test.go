package domain

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func withFeederID(d map[string]any) {
	for _, kind := range listAt(d, "kinds") {
		for _, effector := range kind.(map[string]any)["effectors"].([]any) {
			if schema := effector.(map[string]any)["args_schema"].(map[string]any); effector.(map[string]any)["name"] == "bulk_feeder__set_feed" {
				schema["properties"].(map[string]any)["feeder_id"] = map[string]any{"type": "string"}
				schema["required"] = append(schema["required"].([]any), "feeder_id")
			}
		}
	}
}

func policy(keys ...any) []any { return append([]any{"ground_truth", "baseline_policies", 0}, keys...) }

func TestEveryBaselinePolicyNamesSomethingTheDomainDeclares(t *testing.T) {
	t.Parallel()
	const at = "/ground_truth/baseline_policies/0"
	runRefusals(t, testsupport.BeltConveyorDomain("a"), []refusalCase{
		{"a condition on an unknown kind", func(d map[string]any) { put(d, "motor", policy("when", "kind")...) },
			at + "/when/kind", `baseline policy "b1" watches kind "motor", which the domain does not declare`},
		{"a condition on a state the kind lacks", func(d map[string]any) { put(d, "delivered", policy("when", "state")...) },
			at + "/when/state", `baseline policy "b1" watches "delivered", which is not a state of kind "conveyor_drive"`},
		{"a call of an unknown effector", func(d map[string]any) { put(d, "set_feed", policy("then", 0, "effector")...) },
			at + "/then/0/effector", `the domain declares no effector "set_feed"`},
		{"a target that is no instance", func(d map[string]any) { put(d, "F-9", policy("then", 0, "target")...) },
			at + "/then/0/target", `target "F-9" is neither matched nor a declared instance name`},
		{"a named target whose kind lacks the effector", func(d map[string]any) { put(d, "D-1", policy("then", 0, "target")...) },
			at + "/then/0/effector", `the call goes to a "conveyor_drive" instance, whose kind does not declare effector "bulk_feeder__set_feed"`},
		{"a matched target whose kind lacks the effector", func(d map[string]any) { put(d, model.PolicyTargetMatched, policy("then", 0, "target")...) },
			at + "/then/0/effector", `the call goes to a "conveyor_drive" instance, whose kind does not declare effector "bulk_feeder__set_feed"`},
		{"an argument the effector does not declare", func(d map[string]any) { put(d, 5.0, policy("then", 0, "args", "rate")...) },
			at + "/then/0/args/rate", `effector "bulk_feeder__set_feed" declares no argument "rate"`},
		{"the entity argument given", func(d map[string]any) {
			withFeederID(d)
			put(d, "F-1", policy("then", 0, "args", "feeder_id")...)
		},
			at + "/then/0/args/feeder_id", `argument "feeder_id" of "bulk_feeder__set_feed" is the call's entity, which its target supplies`},
		{"an argument outside its schema", func(d map[string]any) { put(d, 120.0, policy("then", 0, "args", "percent")...) },
			at + "/then/0/args/percent", `argument "percent" of "bulk_feeder__set_feed": `},
		{"a required argument left out", func(d map[string]any) { delete(objectAt(d, policy("then", 0)...), "args") },
			at + "/then/0/args", `the call omits argument "percent", which effector "bulk_feeder__set_feed" requires`},
		{"a policy declared twice", func(d map[string]any) {
			list := listAt(d, "ground_truth", "baseline_policies")
			put(d, append(list, list[0]), "ground_truth", "baseline_policies")
		}, "/ground_truth/baseline_policies/4/name", `baseline policy "b1" is declared twice`},
	})
}

func TestTheBaselinePoliciesHaveClosedShapes(t *testing.T) {
	t.Parallel()
	const at = "/ground_truth/baseline_policies/0"
	runRefusals(t, testsupport.BeltConveyorDomain("a"), []refusalCase{
		{"an operator outside the closed set", func(d map[string]any) { put(d, "approx", policy("when", "operator")...) }, at + "/when/operator", ""},
		{"a condition with an expression", func(d map[string]any) { put(d, "available < 0.5", policy("when", "expr")...) }, at + "/when/expr", ""},
		{"a policy with no call", func(d map[string]any) { put(d, []any{}, policy("then")...) }, at + "/then", ""},
		{"a negative hold", func(d map[string]any) { put(d, -1.0, policy("for_s")...) }, at + "/for_s", ""},
		{"a negative delay", func(d map[string]any) { put(d, -1.0, policy("then", 0, "delay_s")...) }, at + "/then/0/delay_s", ""},
		{"a call without a target", func(d map[string]any) { delete(objectAt(d, policy("then", 0)...), "target") }, at + "/then/0/target", ""},
	})
}

func TestAMatchedCallGoesToTheKindWhoseConditionFired(t *testing.T) {
	t.Parallel()
	path := testsupport.BeltConveyorDomain("a")
	raw := documentOf(t, path, func(d map[string]any) {
		put(d, []any{map[string]any{"effector": "conveyor_drive__stop_drive", "target": "matched", "delay_s": 5.0, "args": map[string]any{"run_command": 0.0}}}, policy("then")...)
	})
	c, err := Parse(raw, path)
	if err != nil {
		t.Fatal(err)
	}
	call := c.Spec.GroundTruth.BaselinePolicies[0].Then[0]
	if !call.TargetsMatched() || call.DelayS != 5 || len(call.Args) != 1 || call.Args["run_command"] != 0.0 {
		t.Fatalf("call = %+v, want a matched stop_drive to 0 5 s after the firing", call)
	}
}

// conveyorBaselines are B1-B4 of PLAN.md section 2.2 as decided in O27: b2's threshold is 1.1 times the drive current at
// full feed on three drives, b3 feeds the plant's safe limp feed. Each carries the label a report shows beside its name.
func conveyorBaselines(nominalAmps, limpPct float64) []model.BaselinePolicy {
	feed := func(percent float64) model.PolicyCall {
		return model.PolicyCall{Effector: "bulk_feeder__set_feed", Target: "F-1", Args: map[string]any{"percent": percent}}
	}
	stop := model.PolicyCall{Effector: "conveyor_drive__stop_drive", Target: model.PolicyTargetMatched, Args: map[string]any{"run_command": 0.0}, DelayS: 60}
	return []model.BaselinePolicy{
		{Name: "b1", Label: "vibration alarm: feed 0, then stop the drive", When: model.PolicyCondition{Kind: "conveyor_drive", State: "vibration", Operator: "gt", Threshold: 4.5}, ForS: 60, Then: []model.PolicyCall{feed(0), stop}},
		{Name: "b2", Label: "current threshold: feed 67 %", When: model.PolicyCondition{Kind: "conveyor_drive", State: "current", Operator: "gt", Threshold: nominalAmps}, ForS: 60, Then: []model.PolicyCall{feed(67)}},
		{Name: "b3", Label: fmt.Sprintf("safe limp feed: %.0f %% from the start", limpPct), When: model.PolicyCondition{Kind: "bulk_feeder", State: "feed_set", Operator: "gte", Threshold: 0}, Then: []model.PolicyCall{feed(limpPct)}},
		{Name: "b4", Label: "drive lost: feed 67 %", When: model.PolicyCondition{Kind: "conveyor_drive", State: "available", Operator: "lt", Threshold: 0.5}, Then: []model.PolicyCall{feed(67)}},
	}
}

func TestBothConveyorPlantsCarryTheirOutcomeWeightsAndBaselinesB1ToB4(t *testing.T) {
	t.Parallel()
	for regime, want := range map[string][]model.BaselinePolicy{"a": conveyorBaselines(3.385, 54), "b": conveyorBaselines(4.688, 18)} {
		c, err := Load(testsupport.BeltConveyorDomain(regime))
		if err != nil {
			t.Fatal(err)
		}
		truth := c.Spec.GroundTruth
		if !reflect.DeepEqual(truth.BaselinePolicies, want) {
			t.Fatalf("%s: baseline policies = %+v\nwant %+v", regime, truth.BaselinePolicies, want)
		}
		w := truth.OutcomeWeights
		if w == nil || len(w.Boundaries) != 1 || w.Boundaries[0].PerViolation != 4 || len(w.Trips) != 5 || len(w.KPIs) != 3 || len(w.Calls) != 3 || w.Tie != 0.1 || w.Unit != "equivalent production hours" {
			t.Fatalf("%s: outcome weights = %+v", regime, w)
		}
	}
}
