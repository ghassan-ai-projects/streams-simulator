package domain

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
)

func TestKindStemsBecomeTheKindFilesOfTheWorkspaceInOrderWithoutRepeats(t *testing.T) {
	t.Parallel()
	suggest, err := Lookup("kind.suggest")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		kinds any
		want  []ArgvPart
	}{
		{"one kind", []any{"fan"}, []ArgvPart{{Path: "kinds/fan.kind.json", Existing: true}}},
		{"a list of text", []string{"fan", "thermal_sensor"}, []ArgvPart{{Path: "kinds/fan.kind.json", Existing: true}, {Path: "kinds/thermal_sensor.kind.json", Existing: true}}},
		{"a kind named twice", []any{"thermal_sensor", "fan", "thermal_sensor"}, []ArgvPart{{Path: "kinds/thermal_sensor.kind.json", Existing: true}, {Path: "kinds/fan.kind.json", Existing: true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := suggest.Argv(map[string]any{"kinds": tc.kinds})
			head := []ArgvPart{{Literal: "kind"}, {Literal: "twin"}, {Literal: "--suggest"}}
			if want := append(head, tc.want...); err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("Argv = %+v, %v, want %+v", got, err, want)
			}
		})
	}
}

func TestKindStemsRefuseWhatIsNotTheNameOfAKindFile(t *testing.T) {
	t.Parallel()
	suggest, err := Lookup("kind.suggest")
	if err != nil {
		t.Fatal(err)
	}
	tooMany := make([]any, maxStems+1)
	for i := range tooMany {
		tooMany[i] = "kind" + strconv.Itoa(i)
	}
	cases := []struct {
		name  string
		kinds any
	}{
		{"a path", []any{"kinds/fan.kind.json"}},
		{"a folder", []any{"other/fan"}},
		{"a parent", []any{".."}},
		{"a flag", []any{"-fan"}},
		{"an empty name", []any{""}},
		{"no kind", []any{}},
		{"a number", []any{4}},
		{"text instead of a list", "fan"},
		{"more kinds than a board holds", tooMany},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got, err := suggest.Argv(map[string]any{"kinds": tc.kinds}); !errors.Is(err, ErrBadArguments) {
				t.Fatalf("Argv = %+v, %v, want ErrBadArguments", got, err)
			}
		})
	}
}

func TestComponentListTakesTheReadingRangeAsNumbers(t *testing.T) {
	t.Parallel()
	list, err := Lookup("component.list")
	if err != nil {
		t.Fatal(err)
	}
	got, err := list.Argv(map[string]any{"unit": "Pa", "min": -5.0, "max": 100000.0})
	want := []ArgvPart{{Literal: "component"}, {Literal: "list"}, {Literal: "--unit"}, {Literal: "Pa"}, {Literal: "--min"}, {Literal: "-5"}, {Literal: "--max"}, {Literal: "100000"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v, want %+v", got, err, want)
	}
	if _, err := list.Argv(map[string]any{"unit": "Pa", "min": "0"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a minimum that is text = %v, want ErrBadArguments", err)
	}
}
