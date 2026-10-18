package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestOnlyKindCheckCarriesItsContentOnStandardInput(t *testing.T) {
	t.Parallel()
	for _, action := range Actions() {
		if (action.Input != "") != (action.Name == "kind.check") {
			t.Errorf("action %s has input %q: only kind.check reads standard input", action.Name, action.Input)
		}
	}
	check, err := Lookup("kind.check")
	if err != nil || check.Mutating || check.Etag || check.Input != "content" {
		t.Fatalf("kind.check = %+v, %v, want a read-only action whose content goes on standard input", check, err)
	}
	listed, err := json.Marshal(check)
	if err != nil || !strings.Contains(string(listed), `"input":"content"`) {
		t.Fatalf("listed kind.check = %s, %v, want it to name its input", listed, err)
	}
}

func TestStdinTakesTheTextOfTheInputArgumentUpToTheLimit(t *testing.T) {
	t.Parallel()
	check, _ := Lookup("kind.check")
	cases := []struct {
		name string
		args map[string]any
		want []byte
		err  error
	}{
		{"text", map[string]any{"content": "{}"}, []byte("{}"), nil},
		{"empty text", map[string]any{"content": ""}, []byte{}, nil},
		{"text at the limit", map[string]any{"content": strings.Repeat("a", MaxInputBytes)}, []byte(strings.Repeat("a", MaxInputBytes)), nil},
		{"text over the limit", map[string]any{"content": strings.Repeat("a", MaxInputBytes+1)}, nil, ErrTooLarge},
		{"absent", map[string]any{}, nil, ErrBadArguments},
		{"a number", map[string]any{"content": 1.0}, nil, ErrBadArguments},
		{"an object", map[string]any{"content": map[string]any{"kind": "k"}}, nil, ErrBadArguments},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := check.Stdin(tc.args)
			if !errors.Is(err, tc.err) || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Stdin = %d bytes, %v, want %d bytes and %v", len(got), err, len(tc.want), tc.err)
			}
		})
	}
	validate, _ := Lookup("kind.validate")
	if got, err := validate.Stdin(map[string]any{"content": "{}"}); got != nil || err != nil {
		t.Fatalf("an action without input = %v, %v, want no standard input", got, err)
	}
}

func TestKindCheckArgvIsFixedAndTheContentIsNeverAnArgument(t *testing.T) {
	t.Parallel()
	check, _ := Lookup("kind.check")
	cases := []struct {
		name string
		args map[string]any
		want []ArgvPart
	}{
		{"plant wire schema only", map[string]any{"content": "kinds/x.kind.json"}, []ArgvPart{{Literal: "kind"}, {Literal: "validate"}, {Literal: "--stdin"}}},
		{"with the twin requirements", map[string]any{"content": "{}", "twin": true}, []ArgvPart{{Literal: "kind"}, {Literal: "validate"}, {Literal: "--stdin"}, {Literal: "--twin"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got, err := check.Argv(tc.args); err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Argv = %+v, %v, want %+v", got, err, tc.want)
			}
		})
	}
	if _, err := check.Argv(map[string]any{"content": "{}", "path": "kinds/x.kind.json"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a path argument = %v, want bad arguments", err)
	}
}

func TestKindSuggestAndComponentListUnitAreRegisteredWithTheirArguments(t *testing.T) {
	t.Parallel()
	suggest, err := Lookup("kind.suggest")
	if err != nil || suggest.Mutating || !suggest.Library || suggest.Kinds || len(suggest.CLIArgs) != 1 || suggest.CLIArgs[0].Shape != ShapeStems {
		t.Fatalf("kind.suggest = %+v, %v, want a read-only action over the library and the named kind files, never the kinds folder", suggest, err)
	}
	list, _ := Lookup("component.list")
	got, err := list.Argv(map[string]any{"unit": "Cel"})
	want := []ArgvPart{{Literal: "component"}, {Literal: "list"}, {Literal: "--unit"}, {Literal: "Cel"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("component.list Argv = %+v, %v, want %+v", got, err, want)
	}
}
