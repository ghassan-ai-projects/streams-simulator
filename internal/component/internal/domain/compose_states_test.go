package domain

import "testing"

func vocabularyOf(entries ...map[string]any) []any {
	list := make([]any, len(entries))
	for i, entry := range entries {
		list[i] = entry
	}
	return list
}

func codedState(value int, name string, attention any) map[string]any {
	return map[string]any{"value": value, "name": name, "attention": attention}
}

func TestComposeCopiesAnInstrumentsStateVocabularyOntoItsChannelOnly(t *testing.T) {
	t.Parallel()
	states := vocabularyOf(codedState(0, "idle", nil), codedState(1, "fault", "alarm"), codedState(2, "degraded", "warn"))
	variant := instrumentVariant(t, "acme/probe-coded", "degC", func(doc map[string]any) { doc["states"] = states })
	doc := composeOK(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-coded@1")))
	got, _ := channelOf(t, doc, "temp")["states"].([]any)
	if len(got) != 3 {
		t.Fatalf("temp channel states = %v, want the instrument's three", got)
	}
	for i, want := range states {
		entry, _ := got[i].(map[string]any)
		declared := want.(map[string]any)
		if entry["name"] != declared["name"] || entry["attention"] != declared["attention"] || number(t, entry["value"]) != float64(i) {
			t.Errorf("state %d = %v, want %v", i, entry, declared)
		}
	}
	if _, present := channelOf(t, doc, "power")["states"]; present {
		t.Error("the power channel carries states although its instrument declares none")
	}
}

func TestComposeWritesNoStatesForAnInstrumentWithoutAVocabulary(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	if _, present := channelOf(t, doc, "temp")["states"]; present {
		t.Error("a channel carries states although its instrument declares none")
	}
}

func TestLoadRefusesAnInstrumentVocabularyThatBreaksItsRules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		states                 []any
	}{
		{"a value declared twice", "/states/1/value", "state value 3 is declared twice", vocabularyOf(codedState(3, "a", nil), codedState(3, "b", nil))},
		{"an attention outside alarm and warn", "", "", vocabularyOf(codedState(0, "a", "critical"))},
		{"a state without a name", "", "", vocabularyOf(codedState(0, "", nil))},
		{"an empty vocabulary", "", "", []any{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files := baseFiles()
			spare := mutate(t, probeJSON, func(doc map[string]any) { doc["id"] = "acme/spare"; doc["states"] = tc.states })
			files.instruments = append(files.instruments, file("instruments/spare.instrument.json", spare))
			_, err := LoadLibrary([]Directory{files.directory("lib")})
			if tc.message == "" {
				wantViolation(t, err, "instrument-v0.1")
				return
			}
			wantIssue(t, err, "instrument-v0.1", tc.pointer, tc.message)
		})
	}
}
