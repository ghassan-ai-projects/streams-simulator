package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func withChannelStates(t *testing.T, states any) []byte {
	t.Helper()
	example, err := Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(example.Spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	kind := doc["kinds"].([]any)[0].(map[string]any)
	kind["channels"].([]any)[0].(map[string]any)["states"] = states
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func stateEntry(value int, name string, attention any) map[string]any {
	return map[string]any{"value": value, "name": name, "attention": attention}
}

func TestAChannelMayDeclareItsStateVocabularyAndTheDomainKeepsIt(t *testing.T) {
	t.Parallel()
	states := []any{stateEntry(0, "normal", nil), stateEntry(1, "alarm", "alarm"), stateEntry(2, "watch", "warn")}
	compiled, err := Parse(withChannelStates(t, states), "states.json")
	if err != nil {
		t.Fatal(err)
	}
	got := compiled.Spec.Kinds[0].Channels[0].States
	if len(got) != 3 || got[0].Attention != nil || *got[1].Attention != model.AttentionAlarm || *got[2].Attention != model.AttentionWarn || got[2].Name != "watch" || got[2].Value != 2 {
		t.Fatalf("states = %+v, want the three declared, with attention kept", got)
	}
}

func TestAChannelVocabularyThatBreaksItsRulesIsRefusedAtItsPointer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		states                 any
	}{
		{"a value declared twice", "/kinds/0/channels/0/states/1/value", "declares state value 4 twice", []any{stateEntry(4, "a", nil), stateEntry(4, "b", nil)}},
		{"an attention outside alarm and warn", "", "", []any{stateEntry(0, "a", "critical")}},
		{"a fractional value", "", "", []any{stateEntry(0, "a", nil), map[string]any{"value": 1.5, "name": "b"}}},
		{"a state without a name", "", "", []any{stateEntry(0, "", nil)}},
		{"an empty vocabulary", "", "", []any{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(withChannelStates(t, tc.states), "states.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) == 0 {
				t.Fatalf("err = %v, want a validation error", err)
			}
			if tc.pointer != "" && (invalid.Issues[0].Pointer != tc.pointer || !strings.Contains(invalid.Issues[0].Message, tc.message)) {
				t.Fatalf("issue = %+v, want %s %q", invalid.Issues[0], tc.pointer, tc.message)
			}
		})
	}
}
