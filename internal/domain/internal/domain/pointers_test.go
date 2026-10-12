package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func mutated(t *testing.T, change func(kind, instances, doc map[string]any)) []byte {
	t.Helper()
	c, err := Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(c.Raw, &doc); err != nil {
		t.Fatal(err)
	}
	kind := doc["kinds"].([]any)[0].(map[string]any)
	instances := doc["instances"].([]any)[0].(map[string]any)
	change(kind, instances, doc)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestEveryCrossReferenceIssueCarriesAPointerIntoTheV02Document(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		change  func(kind, instances, doc map[string]any)
		pointer string
	}{
		{"channel observes", func(kind, _, _ map[string]any) { kind["channels"].([]any)[1].(map[string]any)["observes"] = "nope" }, "/kinds/0/channels/1/observes"},
		{"fault state", func(kind, _, _ map[string]any) {
			kind["faults"].([]any)[2].(map[string]any)["affects"].([]any)[0].(map[string]any)["state"] = "nope"
		}, "/kinds/0/faults/2/affects/0/state"},
		{"fault effector", func(kind, _, _ map[string]any) {
			kind["faults"].([]any)[0].(map[string]any)["expected_effector"] = "nope"
		}, "/kinds/0/faults/0/expected_effector"},
		{"effector state", func(kind, _, _ map[string]any) {
			kind["effectors"].([]any)[0].(map[string]any)["effect"].(map[string]any)["state_deltas"].([]any)[0].(map[string]any)["state"] = "nope"
		}, "/kinds/0/effectors/0/effect/state_deltas/0/state"},
		{"dynamics target", func(kind, _, _ map[string]any) { kind["dynamics"].([]any)[0].(map[string]any)["target"] = "nope" }, "/kinds/0/dynamics/0/target"},
		{"profile weight", func(_, _, doc map[string]any) {
			doc["profiles"].([]any)[0].(map[string]any)["fault_weights"] = map[string]any{"nope/x": 1.0}
		}, "/profiles/0/fault_weights/nope~1x"},
		{"duplicate state", func(kind, _, _ map[string]any) {
			states := kind["state"].([]any)
			kind["state"] = append(states, states[0])
		}, "/kinds/0/state/10/name"},
		{"instances kind", func(_, instances, _ map[string]any) { instances["kind"] = "other" }, "/instances/0/kind"},
		{"count default", func(_, instances, _ map[string]any) { instances["count"].(map[string]any)["default"] = 100.0 }, "/instances/0/count/default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(mutated(t, tc.change), "mutated.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || !invalid.Issues[0].Located || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("err = %v (%#v), want one issue at %s", err, invalid.Issues, tc.pointer)
			}
		})
	}
}

func TestAMissingRequiredMemberIsReportedAtTheMember(t *testing.T) {
	t.Parallel()
	raw := mutated(t, func(kind, _, _ map[string]any) { delete(kind["channels"].([]any)[0].(map[string]any), "resolution") })
	_, err := Parse(raw, "missing.json")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kinds/0/channels/0/resolution" {
		t.Fatalf("err = %v, want one issue at the missing resolution member", err)
	}
}
