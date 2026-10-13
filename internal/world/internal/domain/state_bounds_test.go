package domain

import "testing"

func ptr(v float64) *float64 { return &v }

func TestAStateBoundClampsOnlyTheSideItNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		lower, upper *float64
		in, want     float64
	}{
		{"no bounds", nil, nil, -7, -7},
		{"a lower bound only, below", ptr(5), nil, 2, 5},
		{"a lower bound only, above", ptr(5), nil, 900, 900},
		{"an upper bound only, below", nil, ptr(-3), -10, -10},
		{"an upper bound only, above", nil, ptr(-3), 4, -3},
		{"both, zero width", ptr(0), ptr(0), 4, 0},
		{"both", ptr(1), ptr(2), 1.5, 1.5},
	}
	for _, tc := range cases {
		if got := clampToBounds(tc.in, tc.lower, tc.upper); got != tc.want {
			t.Fatalf("%s: clamp(%v) = %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}
