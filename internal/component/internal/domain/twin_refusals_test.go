package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func chosenKind(doc map[string]any) map[string]any { return doc["kinds"].([]any)[0].(map[string]any) }

func chosen(kind map[string]any, section, name string) map[string]any {
	return kind[section].(map[string]any)[name].(map[string]any)
}

func sectionOf(kind map[string]any, section string) map[string]any {
	return kind[section].(map[string]any)
}

func heaterChoices(t *testing.T, change func(kind map[string]any)) string {
	t.Helper()
	return mutate(t, heaterChoicesJSON, func(doc map[string]any) { change(chosenKind(doc)) })
}

func fileMetric(doc map[string]any) map[string]any { return doc["metrics"].([]any)[0].(map[string]any) }

func fileCommand(doc map[string]any) map[string]any {
	return doc["commands"].([]any)[0].(map[string]any)
}

func wantTwinRefusal(t *testing.T, kinds map[string]map[string]any, choices string, sentinel error, pointerWant string) {
	t.Helper()
	_, err := ScaffoldTwin(baseLibrary(t), kinds, []byte(choices), "test.twin.json")
	invalid := asViolation(t, err, model.ContractTwinChoices)
	if !errors.Is(err, sentinel) || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != pointerWant {
		t.Fatalf("ScaffoldTwin = %v (issues %+v), want %v at %s", err, invalid.Issues, sentinel, pointerWant)
	}
}

func TestAKindFileThatHoldsWhatATwinCannotReproduceIsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		change func(doc map[string]any)
	}{
		{"a metric member the interface does not carry", func(doc map[string]any) { fileMetric(doc)["description"] = "text" }},
		{"a metric without its writable flag", func(doc map[string]any) { delete(fileMetric(doc), "writable") }},
		{"a command without a risk class", func(doc map[string]any) { delete(fileCommand(doc), "risk_class") }},
		{"a critical command", func(doc map[string]any) { fileCommand(doc)["risk_class"] = "critical" }},
		{"an enum metric", func(doc map[string]any) { fileMetric(doc)["data_type"] = "enum" }},
		{"a number metric without a unit", func(doc map[string]any) { delete(fileMetric(doc), "unit") }},
		{"a kind without a metric", func(doc map[string]any) { doc["metrics"] = []any{} }},
		{"a metric name beyond the limit", func(doc map[string]any) { fileMetric(doc)["name"] = "temperature_of_the_heating_element" }},
		{"a member beyond the interface", func(doc map[string]any) { doc["description"] = "text" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantTwinRefusal(t, heaterKinds(t, tc.change), heaterChoicesJSON, ErrKindNotReproducible, "/kinds/0/kind")
		})
	}
}

func TestChoicesThatDoNotFitTheKindFilesAreRefusedAtTheMemberThatIsWrong(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		change   func(kind map[string]any)
		sentinel error
		pointer  string
	}{
		{"a kind no kind file defines", func(k map[string]any) { k["kind"] = "ghost" }, ErrNoKindFile, "/kinds/0/kind"},
		{"a metric left undecided", func(k map[string]any) { delete(sectionOf(k, "metrics"), "power") }, ErrIncompleteChoices, "/kinds/0/metrics"},
		{"a metric the kind file does not hold", func(k map[string]any) { sectionOf(k, "metrics")["ghost"] = map[string]any{"instrument": "none"} }, ErrUnknownChoice, "/kinds/0/metrics/ghost"},
		{"a command left undecided", func(k map[string]any) { delete(sectionOf(k, "commands"), "ping") }, ErrIncompleteChoices, "/kinds/0/commands"},
		{"a response for a safe stop", func(k map[string]any) { sectionOf(k, "commands")["halt"] = map[string]any{"response": "none"} }, ErrUnknownChoice, "/kinds/0/commands/halt"},
		{"an initial value outside the range", func(k map[string]any) { chosen(k, "metrics", "temperature")["initial"] = 500 }, ErrValueOutOfRange, "/kinds/0/metrics/temperature/initial"},
		{"an instrument the library lacks", func(k map[string]any) { chosen(k, "metrics", "temperature")["instrument"] = "acme/ghost" }, ErrInstrumentMismatch, "/kinds/0/metrics/temperature/instrument"},
		{"equipment where an instrument is wanted", func(k map[string]any) { chosen(k, "metrics", "temperature")["instrument"] = "acme/kiln" }, ErrInstrumentMismatch, "/kinds/0/metrics/temperature/instrument"},
		{"an instrument of another quantity", func(k map[string]any) { chosen(k, "metrics", "temperature")["instrument"] = "acme/meter" }, ErrInstrumentMismatch, "/kinds/0/metrics/temperature/instrument"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantTwinRefusal(t, kindDocuments(t, heaterKindJSON), heaterChoices(t, tc.change), tc.sentinel, tc.pointer)
		})
	}
}

func booleanTemperature(doc map[string]any) {
	metric := fileMetric(doc)
	metric["data_type"] = "boolean"
	delete(metric, "unit")
}

func TestAnInstrumentThatReadsLessThanTheMetricOrInAnotherUnitIsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		change func(doc map[string]any)
	}{
		{"a range wider than the instrument", func(doc map[string]any) { fileMetric(doc)["eu_range"] = map[string]any{"min": 0, "max": 500} }},
		{"a unit that needs a conversion", func(doc map[string]any) { fileMetric(doc)["unit"] = "degF" }},
		{"a boolean metric", booleanTemperature},
		{"a metric without a range", func(doc map[string]any) { delete(fileMetric(doc), "eu_range") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantTwinRefusal(t, heaterKinds(t, tc.change), heaterChoicesJSON, ErrInstrumentMismatch, "/kinds/0/metrics/temperature/instrument")
		})
	}
}
