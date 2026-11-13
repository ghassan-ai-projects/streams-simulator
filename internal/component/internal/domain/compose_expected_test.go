package domain

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func expecting(actions map[string]any) map[string]any {
	return kilnEntry(map[string]any{"as": "kiln_a", "expected_actions": actions})
}

func TestComposeWritesAnEntrysExpectedActionsOntoItsQualifiedFault(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, expecting(map[string]any{
		"fouling":      []any{map[string]any{"kind": "kiln_a", "effector": "set_setpoint"}},
		"temp__offset": []any{map[string]any{"kind": "kiln_a", "effector": "set_setpoint"}},
	})))
	faults := member(t, doc, "kinds", 0, "faults")
	want := []any{map[string]any{"kind": "kiln_a", "effector": "kiln_a__set_setpoint"}}
	for _, id := range []string{"kiln_a__fouling", "kiln_a__temp__offset"} {
		if got := named(t, faults, "fault", "id", id)["expected_actions"]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s expected_actions = %v, want %v", id, got, want)
		}
	}
	if _, declared := named(t, faults, "fault", "id", "kiln_a__power__offset")["expected_actions"]; declared {
		t.Error("a fault the entry names no action for got expected actions")
	}
}

func TestComposeRefusesExpectedActionsForAFaultTheEquipmentDoesNotDeclare(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, expecting(map[string]any{
		"melting": []any{map[string]any{"kind": "kiln_a", "effector": "set_setpoint"}},
	})))
	wantIssue(t, err, model.ContractComposition, "/equipment/0/expected_actions/melting", `declares no fault "melting"`)
}

func TestComposeLeavesAnExpectedActionOfAnUndeclaredEffectorToTheDomainRules(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, expecting(map[string]any{
		"fouling": []any{map[string]any{"kind": "kiln_a", "effector": "open_door"}},
	})))
	wantIssue(t, err, model.ContractDomainSpec, "/kinds/0/faults/0/expected_actions/0/effector", `expects effector "kiln_a__open_door"`)
}
