package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type refusalCase struct {
	name    string
	change  func(doc map[string]any)
	pointer string
	message string
}

func runRefusals(t *testing.T, path string, cases []refusalCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(documentOf(t, path, tc.change), path)
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("err = %v, want an issue at %s", err, tc.pointer)
			}
			if !strings.Contains(invalid.Issues[0].Message, tc.message) {
				t.Fatalf("message = %q, want it to say %q", invalid.Issues[0].Message, tc.message)
			}
		})
	}
}

func weights(keys ...any) []any { return append([]any{"ground_truth", "outcome_weights"}, keys...) }

func TestEveryOutcomeWeightNamesSomethingTheDomainDeclares(t *testing.T) {
	t.Parallel()
	const at = "/ground_truth/outcome_weights"
	runRefusals(t, testsupport.BeltConveyorDomain("a"), []refusalCase{
		{"a boundary of an unknown kind", func(d map[string]any) { put(d, "pump", weights("boundaries", 0, "kind")...) },
			at + "/boundaries/0/kind", `the domain declares no kind "pump"`},
		{"a boundary the kind does not declare", func(d map[string]any) { put(d, "spilled", weights("boundaries", 0, "name")...) },
			at + "/boundaries/0/name", `kind "belt_conveyor" declares no boundary "spilled"`},
		{"a boundary weighted twice", func(d map[string]any) {
			list := listAt(d, weights("boundaries")...)
			put(d, append(list, list[0]), weights("boundaries")...)
		}, at + "/boundaries/1/name", `boundary "stranded" of kind "belt_conveyor" is weighted twice`},
		{"a trip on a state that is no trip", func(d map[string]any) { put(d, "winding", weights("trips", 0, "state")...) },
			at + "/trips/0/state", `"winding" is not a trip state of kind "conveyor_drive"`},
		{"a trip of an unknown kind", func(d map[string]any) { put(d, "motor", weights("trips", 1, "kind")...) },
			at + "/trips/1/kind", `the domain declares no kind "motor"`},
		{"a trip weighted twice", func(d map[string]any) { put(d, "relay_ok", weights("trips", 1, "state")...) },
			at + "/trips/1/state", `trip "relay_ok" of kind "conveyor_drive" is weighted twice`},
		{"a KPI that is only a state", func(d map[string]any) { put(d, "winding", weights("kpis", 0, "state")...) },
			at + "/kpis/0/state", `"winding" is not a KPI of kind "conveyor_drive"`},
		{"a KPI of another kind", func(d map[string]any) { put(d, "bulk_feeder", weights("kpis", 2, "kind")...) },
			at + "/kpis/2/state", `"delivered" is not a KPI of kind "bulk_feeder"`},
		{"a KPI weighted twice", func(d map[string]any) { put(d, "life_used", weights("kpis", 1, "state")...) },
			at + "/kpis/1/state", `KPI "life_used" of kind "conveyor_drive" is weighted twice`},
		{"a call of an unknown effector", func(d map[string]any) { put(d, "set_feed", weights("calls", 0, "effector")...) },
			at + "/calls/0/effector", `the domain declares no effector "set_feed"`},
		{"a call weighted twice", func(d map[string]any) { put(d, "bulk_feeder__set_feed", weights("calls", 2, "effector")...) },
			at + "/calls/2/effector", `calls of "bulk_feeder__set_feed" are weighted twice`},
	})
}

func TestTheOutcomeWeightsHaveClosedShapes(t *testing.T) {
	t.Parallel()
	const at = "/ground_truth/outcome_weights"
	runRefusals(t, testsupport.BeltConveyorDomain("a"), []refusalCase{
		{"a boundary without a weight", func(d map[string]any) { delete(objectAt(d, weights("boundaries", 0)...), "per_violation") },
			at + "/boundaries/0", ""},
		{"a trip without its weight", func(d map[string]any) { delete(objectAt(d, weights("trips", 0)...), "per_trip") }, at + "/trips/0/per_trip", ""},
		{"a weight that is no number", func(d map[string]any) { put(d, "high", weights("kpis", 0, "per_unit")...) }, at + "/kpis/0/per_unit", ""},
		{"an unknown term", func(d map[string]any) { put(d, []any{}, weights("restarts")...) }, at + "/restarts", ""},
	})
}

func listAt(doc map[string]any, keys ...any) []any { return at(doc, keys...).([]any) }

func objectAt(doc map[string]any, keys ...any) map[string]any {
	return at(doc, keys...).(map[string]any)
}
