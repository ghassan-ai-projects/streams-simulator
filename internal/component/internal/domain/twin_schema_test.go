package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestChoicesThatBreakTheirSchemaAreRefusedAtTheMemberThatIsWrong(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		change  func(doc map[string]any)
		pointer string
	}{
		{"no kinds", func(doc map[string]any) { delete(doc, "kinds") }, "/kinds"},
		{"a surprise member", func(doc map[string]any) { doc["surprise"] = true }, "/surprise"},
		{"a bad id", func(doc map[string]any) { doc["id"] = "Heater Twin" }, "/id"},
		{"a response outside the templates", func(doc map[string]any) { chosen(chosenKind(doc), "commands", "ping")["response"] = "ramp" }, "/kinds/0/commands/ping/response"},
		{"an instrument that names no library file", func(doc map[string]any) { chosen(chosenKind(doc), "metrics", "power")["instrument"] = "meter" }, "/kinds/0/metrics/power/instrument"},
		{"a time constant of zero", func(doc map[string]any) { firstDrive(chosenKind(doc))["time_constant_s"] = 0 }, "/kinds/0/commands/set_setpoint/drives/0/time_constant_s"},
		{"a metric choice without an instrument", func(doc map[string]any) { delete(chosen(chosenKind(doc), "metrics", "power"), "instrument") }, "/kinds/0/metrics/power/instrument"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ScaffoldTwin(baseLibrary(t), kindDocuments(t, heaterKindJSON), []byte(mutate(t, heaterChoicesJSON, tc.change)), "test.twin.json")
			invalid := asViolation(t, err, model.ContractTwinChoices)
			for _, issue := range invalid.Issues {
				if issue.Pointer == tc.pointer {
					return
				}
			}
			t.Fatalf("issues = %+v, want one at %s", invalid.Issues, tc.pointer)
		})
	}
}

func TestChoicesThatAreNotJSONAreRefusedAsTheTwinContract(t *testing.T) {
	t.Parallel()
	_, err := ScaffoldTwin(baseLibrary(t), kindDocuments(t, heaterKindJSON), []byte("{"), "test.twin.json")
	invalid := asViolation(t, err, model.ContractTwinChoices)
	if invalid.Source != "test.twin.json" {
		t.Fatalf("source = %q, want the file the choices came from", invalid.Source)
	}
}
