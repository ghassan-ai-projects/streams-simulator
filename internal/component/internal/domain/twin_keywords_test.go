package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestEveryRefusalOfTheChoicesNamesTheRuleItBroke(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name            string
		kinds           map[string]map[string]any
		choices         string
		sentinel        error
		keyword, member string
	}{
		{"a kind no kind file defines", kindDocuments(t, heaterKindJSON), heaterChoices(t, func(k map[string]any) { k["kind"] = "ghost" }), ErrNoKindFile, "twin_kind_missing", "kind"},
		{"a metric left undecided", kindDocuments(t, heaterKindJSON), heaterChoices(t, func(k map[string]any) { delete(sectionOf(k, "metrics"), "power") }), ErrIncompleteChoices, "twin_choice_missing", ""},
		{"a metric the kind file does not hold", kindDocuments(t, heaterKindJSON),
			heaterChoices(t, func(k map[string]any) { sectionOf(k, "metrics")["ghost"] = map[string]any{"instrument": "none"} }), ErrUnknownChoice, "twin_unknown_choice", ""},
		{"an initial value outside the range", kindDocuments(t, heaterKindJSON),
			heaterChoices(t, func(k map[string]any) { chosen(k, "metrics", "temperature")["initial"] = 500 }), ErrValueOutOfRange, "twin_initial_range", "initial"},
		{"an instrument the library lacks", kindDocuments(t, heaterKindJSON),
			heaterChoices(t, func(k map[string]any) { chosen(k, "metrics", "temperature")["instrument"] = "acme/ghost" }), ErrInstrumentMismatch, "twin_instrument_mismatch", "instrument"},
		{"a range wider than the instrument", heaterKinds(t, func(doc map[string]any) { fileMetric(doc)["eu_range"] = map[string]any{"min": 0, "max": 500} }), heaterChoicesJSON,
			ErrInstrumentMismatch, "twin_instrument_mismatch", "eu_range"},
		{"a metric without a range", heaterKinds(t, func(doc map[string]any) { delete(fileMetric(doc), "eu_range") }), heaterChoicesJSON,
			ErrInstrumentMismatch, "twin_instrument_mismatch", "eu_range"},
		{"an instrument of another quantity", kindDocuments(t, heaterKindJSON),
			heaterChoices(t, func(k map[string]any) { chosen(k, "metrics", "temperature")["instrument"] = "acme/meter" }), ErrInstrumentMismatch, "twin_instrument_mismatch", "unit"},
		{"a boolean metric", heaterKinds(t, booleanTemperature), heaterChoicesJSON, ErrInstrumentMismatch, "twin_instrument_mismatch", "data_type"},
		{"a response that does not fit", kindDocuments(t, heaterKindJSON),
			heaterChoices(t, func(k map[string]any) { sectionOf(k, "commands")["ping"] = map[string]any{"response": "direct"} }), ErrResponseMismatch, "twin_response_mismatch", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ScaffoldTwin(baseLibrary(t), tc.kinds, []byte(tc.choices), "test.twin.json")
			invalid := asViolation(t, err, model.ContractTwinChoices)
			if !errors.Is(err, tc.sentinel) || len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != tc.keyword || invalid.Issues[0].Member != tc.member {
				t.Fatalf("ScaffoldTwin = %v (issues %+v), want %v with keyword %s and member %q", err, invalid.Issues, tc.sentinel, tc.keyword, tc.member)
			}
		})
	}
}

func TestAKindFileThatATwinCannotReproduceIsRefusedWithEveryViolationAtOnce(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) {
		delete(fileMetric(doc), "resolution")
		metrics := doc["metrics"].([]any)
		delete(metrics[1].(map[string]any), "resolution")
		delete(metrics[2].(map[string]any), "unit")
		delete(fileCommand(doc), "safe_stop")
	})
	_, err := ScaffoldTwin(baseLibrary(t), kinds, []byte(heaterChoicesJSON), "test.twin.json")
	invalid := asViolation(t, err, model.ContractTwinChoices)
	type rule struct{ keyword, member string }
	var got []rule
	for _, issue := range invalid.Issues {
		if issue.Pointer != "/kinds/0/kind" {
			t.Errorf("issue %+v, want every issue at the kind of the choices", issue)
		}
		got = append(got, rule{issue.Keyword, issue.Member})
	}
	want := []rule{{"twin_member_required", "resolution"}, {"twin_member_required", "resolution"}, {"twin_unit_required", "unit"}, {"twin_member_required", "safe_stop"}}
	if !errors.Is(err, ErrKindNotReproducible) || !reflect.DeepEqual(got, want) {
		t.Fatalf("ScaffoldTwin = %v, rules %v, want ErrKindNotReproducible with %v", err, got, want)
	}
}

func TestAnIssueOfAKindFileSaysWhichMemberOfWhichKindFileIsWrong(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) { delete(fileMetric(doc), "writable") })
	_, err := ScaffoldTwin(baseLibrary(t), kinds, []byte(heaterChoicesJSON), "test.twin.json")
	invalid := asViolation(t, err, model.ContractTwinChoices)
	const want = `kind file "heater" at /metrics/0/writable: needs "writable": the interface a twin reproduces always states it`
	if len(invalid.Issues) != 1 || invalid.Issues[0].Message != want {
		t.Fatalf("issues = %+v, want one saying %q", invalid.Issues, want)
	}
}
