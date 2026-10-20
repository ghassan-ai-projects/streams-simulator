package domain

import (
	"encoding/json"
	"testing"
)

func TestAScaledValueRoundsToTheWholeRawValueItIsWithinFloatErrorOf(t *testing.T) {
	t.Parallel()
	if got := exact(4.35 * 100); got != int64(435) {
		t.Errorf("4.35 x 100 = %#v, want 435", got)
	}
	if got := exact(0.07 * 100); got != int64(7) {
		t.Errorf("0.07 x 100 = %#v, want 7", got)
	}
	if got := exact(43.5); got != 43.5 {
		t.Errorf("43.5 = %#v, want it left as it is", got)
	}
}

func TestAWholeRawArgumentNeedsAKindArgumentWhoseStepTimesTheScaleIsWhole(t *testing.T) {
	t.Parallel()
	raw := map[string]any{"duty": map[string]any{"type": "integer", "minimum": 0, "maximum": 1000}}
	cases := map[string]struct {
		kind  map[string]any
		scale string
		fits  bool
	}{
		"an integer level times 10":          {map[string]any{"type": "integer", "minimum": 0, "maximum": 60}, "10", true},
		"a number with a step of 0.01, x100": {map[string]any{"type": "number", "multipleOf": 0.01, "minimum": 0, "maximum": 10}, "100", true},
		"a number with no step":              {map[string]any{"type": "number", "minimum": 0, "maximum": 10}, "100", false},
		"an integer times a half":            {map[string]any{"type": "integer", "minimum": 0, "maximum": 10}, "0.5", false},
	}
	for name, c := range cases {
		fit := commandFit{at: "/c", kind: map[string]any{"level": c.kind}, raw: raw}
		err := fit.scaled("duty", MappedArg{From: "level", Scale: json.Number(c.scale)})
		if (err == nil) != c.fits {
			t.Errorf("%s: %v, want fits %v", name, err, c.fits)
		}
	}
}
