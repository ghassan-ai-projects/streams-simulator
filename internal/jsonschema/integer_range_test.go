package jsonschema

import (
	"encoding/json"
	"math"
	"testing"
)

// An integer is a number with no fractional part, whatever its magnitude: a
// uint64 seed above int64 range is an integer, and 0.5 or an infinity is not.
func TestIntegerTypeAcceptsEveryWholeNumberInRange(t *testing.T) {
	t.Parallel()
	schema, err := Compile(map[string]any{"type": "integer"})
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		value any
		want  bool
	}{
		"uint64 max":      {uint64(math.MaxUint64), true},
		"json number max": {json.Number("18446744073709551615"), true},
		"2^63":            {json.Number("9223372036854775808"), true},
		"negative":        {int64(math.MinInt64), true},
		"large float":     {1e19, true},
		"fraction":        {0.5, false},
		"infinity":        {math.Inf(1), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := len(schema.Validate(tc.value)) == 0; got != tc.want {
				t.Fatalf("integer(%v) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
