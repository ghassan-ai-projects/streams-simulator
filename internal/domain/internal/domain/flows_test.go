package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestFlowsJoinStoresOfTheirKind(t *testing.T) {
	t.Parallel()
	cases := []violationCase{
		{"a flow from a state that is no store", func(d map[string]any) { put(d, "feed", "kinds", 0, "flows", 0, "from") }, "/kinds/0/flows/0/from"},
		{"a flow to a state that is no store", func(d map[string]any) { put(d, "pump_speed", "kinds", 0, "flows", 0, "to") }, "/kinds/0/flows/0/to"},
		{"a flow to its own source", func(d map[string]any) { put(d, "upper", "kinds", 0, "flows", 0, "to") }, "/kinds/0/flows/0/to"},
		{"an undeclared scale state", func(d map[string]any) { put(d, "valve", "kinds", 0, "flows", 0, "scale", 1) }, "/kinds/0/flows/0/scale/1"},
		{"a flow declared twice", func(d map[string]any) { put(d, "transfer", "kinds", 0, "flows", 1, "name") }, "/kinds/0/flows/1/name"},
		{"stores stepping at different rates", func(d map[string]any) { put(d, 500, "kinds", 0, "dynamics", 2, "dt_ms") }, "/kinds/0/dynamics/2/dt_ms"},
		{"an override store stepping at another rate", func(d map[string]any) {
			put(d, map[string]any{
				"state":    []any{map[string]any{"name": "spare", "initial": 0.0}},
				"dynamics": []any{map[string]any{"target": "spare", "tier": "F1", "dt_ms": 500, "f1": map[string]any{"form": "store", "capacity": 1.0}}},
			}, "instances", 0, "override")
		}, "/instances/0/override/dynamics/0/dt_ms"},
		{"a store without a capacity", func(d map[string]any) { delete(at(d, "kinds", 0, "dynamics", 2, "f1").(map[string]any), "capacity") }, "/kinds/0/dynamics/2/f1/capacity"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(documentOf(t, testsupport.TransferDomain(), tc.change), "transfer.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("err = %v, want an issue at %s", err, tc.pointer)
			}
		})
	}
}

func TestTheTransferDomainKeepsItsStoresAndFlows(t *testing.T) {
	t.Parallel()
	c, err := Load(testsupport.TransferDomain())
	if err != nil {
		t.Fatal(err)
	}
	kind := &c.Spec.Kinds[0]
	if len(kind.Flows) != 2 || !kind.IsStore("upper") || !kind.IsStore("lower") || kind.IsStore("feed") || kind.Flows[1].To != "" {
		t.Fatalf("kind = %+v", kind)
	}
}
