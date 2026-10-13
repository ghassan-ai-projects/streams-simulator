package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func coupledViolation(t *testing.T, change func(doc map[string]any), pointer string) {
	t.Helper()
	_, err := Parse(documentOf(t, testsupport.CoupledDomain(), change), "coupled.json")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != pointer {
		t.Fatalf("want one issue at %s, got %v", pointer, err)
	}
}

func addLinkType(d map[string]any, name, from, to string) {
	put(d, append(d["link_types"].([]any), map[string]any{"name": name, "from": from, "to": to}), "link_types")
}

func linkInput(state, linkType, aggregate string) map[string]any {
	return map[string]any{"state": state, "coef": 1.0, "link": map[string]any{"type": linkType, "aggregate": aggregate}}
}

func TestLinkReadsNameADeclaredLinkIntoTheirKindAndAStateOfItsSource(t *testing.T) {
	t.Parallel()
	cases := []violationCase{
		{"a follow over an undeclared link type", func(d map[string]any) { put(d, "nope", "kinds", 1, "dynamics", 0, "follow", "link") }, "/kinds/1/dynamics/0/follow/link"},
		{"a follow of a state the source lacks", func(d map[string]any) { put(d, "temp", "kinds", 1, "dynamics", 0, "follow", "state") }, "/kinds/1/dynamics/0/follow/state"},
		{"an input of a state the source lacks", func(d map[string]any) { put(d, "temp", "kinds", 1, "dynamics", 3, "f1", "inputs", 0, "state") }, "/kinds/1/dynamics/3/f1/inputs/0/state"},
		{"an input over a link out of the kind", func(d map[string]any) {
			addLinkType(d, "back", "header", "unit")
			put(d, "back", "kinds", 1, "dynamics", 3, "f1", "inputs", 0, "link", "type")
		}, "/kinds/1/dynamics/3/f1/inputs/0/link/type"},
		{"a mean with no default", func(d map[string]any) {
			delete(at(d, "kinds", 1, "dynamics", 3, "f1", "inputs", 0, "link").(map[string]any), "default")
		}, "/kinds/1/dynamics/3/f1/inputs/0/link/default"},
		{"two kinds reading each other", func(d map[string]any) {
			addLinkType(d, "back", "header", "unit")
			put(d, []any{map[string]any{"state": "setpoint", "coef": 1.0}, linkInput("level", "back", "sum")}, "kinds", 0, "dynamics", 1, "f1", "inputs")
		}, "/kinds/0"},
		{"a kind reading itself", func(d map[string]any) {
			addLinkType(d, "peer", "unit", "unit")
			put(d, []any{linkInput("flow", "peer", "sum")}, "kinds", 0, "dynamics", 1, "f1", "inputs")
		}, "/kinds/0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			coupledViolation(t, tc.change, tc.pointer)
		})
	}
}

func TestALinkInputIsNoDependencyInsideItsOwnKind(t *testing.T) {
	t.Parallel()
	_, err := Parse(documentOf(t, testsupport.CoupledDomain(), func(d map[string]any) {
		put(d, append(at(d, "kinds", 1, "state").([]any), map[string]any{"name": "flow", "unit": "m3/h", "initial": 0.0}), "kinds", 1, "state")
		lag := map[string]any{"form": "first_order_lag", "gain": 1.0, "time_constant_s": 10.0, "inputs": []any{"avg_flow"}}
		put(d, append(at(d, "kinds", 1, "dynamics").([]any), map[string]any{"target": "flow", "tier": "F1", "dt_ms": 1000, "f1": lag}), "kinds", 1, "dynamics")
	}), "coupled.json")
	if err != nil {
		t.Fatalf("the header's avg_flow reads the units' flow, not its own: %v", err)
	}
}

func TestAStatefulFormCannotFollowLinkedInstances(t *testing.T) {
	t.Parallel()
	for _, form := range unfollowableForms {
		d := &model.Dynamics{Target: "level", Follow: &model.Follow{Link: "feed", State: "flow", Aggregate: model.AggregateSum}, F1: &model.F1Dyn{Form: form}}
		var invalid *model.ValidationError
		if err := (domainChecks{at: "/kinds/1/dynamics/2"}).dynamicFollow(d); !errors.As(err, &invalid) || invalid.Issues[0].Pointer != "/kinds/1/dynamics/2/follow" {
			t.Fatalf("form %s: %v", form, err)
		}
	}
}

func TestTheCoupledDomainLoads(t *testing.T) {
	t.Parallel()
	c, err := Load(testsupport.CoupledDomain())
	if err != nil || len(c.Spec.Links) != 5 || c.Kind("header").LinkReads() == nil {
		t.Fatalf("coupled domain = %v", err)
	}
}
