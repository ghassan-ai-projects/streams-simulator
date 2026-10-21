package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func thermalMapping(t *testing.T, kind string) MappedKind {
	t.Helper()
	boards, _, err := ParseBoardFiles(semanticSources(t, BoardFileSuffix, asIs))
	if err != nil {
		t.Fatal(err)
	}
	return boards["arduino-thermal"].Kinds[kind]
}

func TestAKindCommandBecomesTheRawCommandTheTwinRuns(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, kind, command string
		args                map[string]any
		raw                 string
		rawArgs             map[string]any
	}{
		{"a level", "fan", "set_cooling", map[string]any{"level_percent": json.Number("45")}, "set_pwm_lease",
			map[string]any{"duty_permille": int64(450), "lease_ms": int64(5000)}},
		{"a safe stop", "fan", "stop_fan", map[string]any{}, "stop_fan", map[string]any{}},
		{"a word", "indicator", "show_state", map[string]any{"state": "watch"}, "set_led",
			map[string]any{"brightness_permille": int64(500), "pattern": int64(1)}},
		{"the last word", "indicator", "show_state", map[string]any{"state": "alert"}, "set_led",
			map[string]any{"brightness_permille": int64(1000), "pattern": int64(2)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			raw, rawArgs, err := RawCommand(thermalMapping(t, c.kind), c.command, c.args)
			if err != nil || raw != c.raw || !reflect.DeepEqual(rawArgs, c.rawArgs) {
				t.Fatalf("RawCommand = %s %#v, %v, want %s %#v", raw, rawArgs, err, c.raw, c.rawArgs)
			}
		})
	}
}

func TestACommandThatDoesNotTranslateIsRefusedBeforeTheTwin(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		kind, command string
		args          map[string]any
	}{
		"an unknown word":    {"indicator", "show_state", map[string]any{"state": "panic"}},
		"a missing level":    {"fan", "set_cooling", map[string]any{}},
		"a raw command name": {"fan", "set_pwm_lease", map[string]any{"duty_permille": 450}},
	} {
		if _, _, err := RawCommand(thermalMapping(t, c.kind), c.command, c.args); !errors.Is(err, ErrNotTranslated) {
			t.Errorf("%s: %v, want ErrNotTranslated", name, err)
		}
	}
	schema := map[string]any{"type": "object", "additionalProperties": false, "required": []any{"level_percent"},
		"properties": map[string]any{"level_percent": map[string]any{"type": "integer", "minimum": 0, "maximum": 60}}}
	for _, args := range []map[string]any{{"level_percent": json.Number("61")}, {"level_percent": 5, "lease_ms": 9000}, {}} {
		if err := CheckArguments(schema, args); !errors.Is(err, ErrNotTranslated) {
			t.Errorf("args %v: %v, want refused by the kind's argument schema", args, err)
		}
	}
	if err := CheckArguments(schema, map[string]any{"level_percent": json.Number("45")}); err != nil {
		t.Errorf("a level within the kind: %v", err)
	}
}

func TestARawReadingBecomesTheKindsMetric(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, kind, channel string
		value               any
		metric              string
		want                any
		carried             bool
	}{
		{"a scaled duty", "fan", "duty_permille", 450.0, "level_percent", 45.0, true},
		{"a noisy pattern code", "indicator", "pattern", 1.00003, "state", "watch", true},
		{"a pattern word", "indicator", "pattern", "solid", "state", "alert", true},
		{"a code no word has", "indicator", "pattern", 7.0, "state", nil, false},
		{"a raw reading with no metric", "indicator", "brightness_permille", 500.0, "", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			metric, value, carried := SemanticReading(thermalMapping(t, c.kind), c.channel, c.value)
			if metric != c.metric || carried != c.carried || (carried && value != c.want) {
				t.Fatalf("SemanticReading = %q %v %v, want %q %v %v", metric, value, carried, c.metric, c.want, c.carried)
			}
		})
	}
}

func TestAnArgumentTheMappingDoesNotUseIsRefusedNotDropped(t *testing.T) {
	t.Parallel()
	for name, args := range map[string]map[string]any{
		"a host lease":     {"level_percent": int64(30), "lease_ms": int64(99999)},
		"an extra member":  {"level_percent": int64(30), "extra": "x"},
		"a safe stop with": {"level_percent": int64(30)},
	} {
		command := "set_cooling"
		if name == "a safe stop with" {
			command = "stop_fan"
		}
		if _, _, err := RawCommand(thermalMapping(t, "fan"), command, args); !errors.Is(err, ErrNotTranslated) {
			t.Errorf("%s: RawCommand = %v, want ErrNotTranslated", name, err)
		}
	}
	if _, _, err := RawCommand(thermalMapping(t, "indicator"), "show_state", map[string]any{"state": "watch", "pattern": "solid"}); !errors.Is(err, ErrNotTranslated) {
		t.Errorf("a raw word beside the kind word: %v, want ErrNotTranslated", err)
	}
}
