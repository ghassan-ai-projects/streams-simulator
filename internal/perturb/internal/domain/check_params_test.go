package domain

import (
	"strings"
	"testing"
)

// CheckApplication states the violation Apply refuses, without the runtime
// prefix.
func TestCheckApplicationStatesWhatApplyRefuses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, perturbation string
		params             map[string]any
		from, until        int64
		want               string
	}{
		{"no params", Drop, nil, 0, 0, ""},
		{"declared rate", Drop, map[string]any{"rate": 0.5}, 0, 0, ""},
		{"window without an end", Drop, nil, 50, 0, ""},
		{"empty window", Drop, nil, 50, 50, ""},
		{"unknown perturbation", "no_such", nil, 0, 0, `unknown perturbation "no_such"`},
		{"window ending before it starts", Drop, nil, 50, 40, "drop window would end at 40, before it starts at 50"},
		{"undeclared parameter", Drop, map[string]any{"bogus": 1.0}, 0, 0, `drop has no parameter "bogus"`},
		{"out of range", Drop, map[string]any{"rate": 2.0}, 0, 0, "drop rate: must be in [0,1], got 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := CheckApplication(tc.perturbation, tc.params, tc.from, tc.until)
			if (tc.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("CheckApplication = %v, want %q", err, tc.want)
			}
			if err != nil && strings.HasPrefix(err.Error(), "perturb:") {
				t.Fatalf("violation %q carries the runtime prefix", err)
			}
		})
	}
}
