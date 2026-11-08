package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func policy(keys ...any) []any { return append([]any{"ground_truth", "baseline_policies", 0}, keys...) }

func TestEveryBaselinePolicyNamesSomethingTheDomainDeclares(t *testing.T) {
	t.Parallel()
	const at = "/ground_truth/baseline_policies/0"
	runRefusals(t, testsupport.BeltConveyorDomain("a"), []refusalCase{
		{"a condition on an unknown kind", func(d map[string]any) { put(d, "motor", policy("when", "kind")...) },
			at + "/when/kind", `baseline policy "b4" watches kind "motor", which the domain does not declare`},
		{"a condition on a state the kind lacks", func(d map[string]any) { put(d, "delivered", policy("when", "state")...) },
			at + "/when/state", `baseline policy "b4" watches "delivered", which is not a state of kind "conveyor_drive"`},
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
		{"the entity argument given", func(d map[string]any) { put(d, "F-1", policy("then", 0, "args", "feeder_id")...) },
			at + "/then/0/args/feeder_id", `argument "feeder_id" of "bulk_feeder__set_feed" is the call's entity, which its target supplies`},
		{"an argument outside its schema", func(d map[string]any) { put(d, 120.0, policy("then", 0, "args", "percent")...) },
			at + "/then/0/args/percent", `argument "percent" of "bulk_feeder__set_feed": `},
		{"a required argument left out", func(d map[string]any) { delete(objectAt(d, policy("then", 0)...), "args") },
			at + "/then/0/args", `the call omits argument "percent", which effector "bulk_feeder__set_feed" requires`},
		{"a policy declared twice", func(d map[string]any) {
			list := listAt(d, "ground_truth", "baseline_policies")
			put(d, append(list, list[0]), "ground_truth", "baseline_policies")
		}, "/ground_truth/baseline_policies/1/name", `baseline policy "b4" is declared twice`},
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
		put(d, []any{map[string]any{"effector": "conveyor_drive__stop_drive", "target": "matched", "delay_s": 5.0}}, policy("then")...)
	})
	c, err := Parse(raw, path)
	if err != nil {
		t.Fatal(err)
	}
	call := c.Spec.GroundTruth.BaselinePolicies[0].Then[0]
	if !call.TargetsMatched() || call.DelayS != 5 || call.Args != nil {
		t.Fatalf("call = %+v, want a matched stop_drive 5 s after the firing", call)
	}
}

func TestBothConveyorPlantsCarryTheirOutcomeWeightsAndPolicyB4(t *testing.T) {
	t.Parallel()
	for _, regime := range []string{"a", "b"} {
		c, err := Load(testsupport.BeltConveyorDomain(regime))
		if err != nil {
			t.Fatal(err)
		}
		truth := c.Spec.GroundTruth
		b4 := truth.BaselinePolicies
		if len(b4) != 1 || b4[0].Name != "b4" || b4[0].When != (model.PolicyCondition{Kind: "conveyor_drive", State: "available", Operator: "lt", Threshold: 0.5}) ||
			len(b4[0].Then) != 1 || b4[0].Then[0].Target != "F-1" || b4[0].Then[0].Args["percent"] != 67.0 {
			t.Fatalf("%s: baseline policies = %+v", regime, b4)
		}
		w := truth.OutcomeWeights
		if w == nil || len(w.Boundaries) != 1 || w.Boundaries[0].PerViolation != 4 || len(w.Trips) != 5 || len(w.KPIs) != 3 || len(w.Calls) != 3 {
			t.Fatalf("%s: outcome weights = %+v", regime, w)
		}
	}
}
