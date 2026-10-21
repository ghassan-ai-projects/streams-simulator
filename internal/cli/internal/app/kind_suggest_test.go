package app

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestKindTwinSuggestPrintsValidChoicesForTheKindFilesAndWritesNothing(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "kind", "twin", "--suggest", "--kinds-dir", testsupport.TwinThermalKindsDir()).mustSucceed(t).document(t, "kind_suggest")
	choices, _ := doc["choices"].(map[string]any)
	kinds, _ := choices["kinds"].([]any)
	if choices["schema_version"] != "0.1" || choices["id"] != "fan-indicator-thermal-sensor-twin" || len(kinds) != 3 || len(doc) != 1 {
		t.Fatalf("kind twin --suggest = %v, want one choices document for the three kinds", doc)
	}
}

func TestKindTwinSuggestRefusesWhatItCannotSuggestFor(t *testing.T) {
	t.Parallel()
	assertFailureCodes(t, []failureCase{
		{"choices beside it", []string{"kind", "twin", "--suggest", "--choices", testsupport.TwinThermalChoices(), "--kinds-dir", testsupport.TwinThermalKindsDir()}, "usage_invalid", 2},
		{"an out file beside it", []string{"kind", "twin", "--suggest", "--out", "x.domain.json", "--kinds-dir", testsupport.TwinThermalKindsDir()}, "usage_invalid", 2},
		{"no kind file", []string{"kind", "twin", "--suggest"}, "kind_failed", 1},
		{"a kind file that is no kind file", []string{"kind", "twin", "--suggest", writeFile(t, "thin.kind.json", `{"kind": "thin"}`)}, "kind_failed", 1},
	})
}

func TestKindTwinSuggestOfAKindATwinCannotReproduceListsEveryViolationAtOnce(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "probe.kind.json", incompleteKind)
	body := invoke(t, "kind", "twin", "--suggest", path).failure(t)
	issues, _ := body["issues"].([]any)
	if body["code"] != "kind_failed" || len(issues) != len(incompleteIssues) {
		t.Fatalf("failure = %v, want kind_failed with the %d issues of the kind file", body, len(incompleteIssues))
	}
}

func TestComponentListByUnitListsInstrumentsExactUnitFirstWithTheirUnitAndRange(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "list", "--unit", "%").mustSucceed(t).document(t, "component_list")
	entries, _ := doc["components"].([]any)
	first, _ := entries[0].(map[string]any)
	second, _ := entries[1].(map[string]any)
	if len(entries) != 2 || first["id"] != "streamsim/valve-position" || first["match"] != "exact" || first["unit"] != "%" ||
		second["id"] != "streamsim/humidity" || second["match"] != "convertible" || second["unit"] != "%RH" || second["range"] == nil {
		t.Fatalf("component list --unit %% = %v, want the valve position first and the humidity probe by conversion", doc)
	}
}

func TestComponentListByAnUnknownUnitIsEmptyAndTheOptionBelongsToListOnly(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "list", "--unit", "furlong").mustSucceed(t).document(t, "component_list")
	if entries, present := doc["components"].([]any); !present || len(entries) != 0 {
		t.Fatalf("component list --unit furlong = %v, want an empty list", doc)
	}
	for _, args := range [][]string{{"component", "show", "streamsim/rtd", "--unit", "Cel"}, {"component", "validate", "--unit", "Cel"}} {
		if got := invoke(t, args...); got.code != 2 || got.failure(t)["code"] != "usage_invalid" {
			t.Errorf("%v: exit %d, want a usage refusal", args, got.code)
		}
	}
}
