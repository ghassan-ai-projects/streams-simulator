package domain

import (
	"math"
	"testing"
)

func TestTheRunSpanIsBoundedWithoutOverflow(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		start, to   int64
		exceedsSpan bool
	}{
		{"within", 0, MaxRunSpanNS, false},
		{"one past", 0, MaxRunSpanNS + 1, true},
		{"backwards", 100, 0, false},
		{"extreme times", math.MinInt64, math.MaxInt64, true},
		{"late start", math.MaxInt64 - 10, math.MaxInt64, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ExceedsRunSpan(tc.start, tc.to); got != tc.exceedsSpan {
				t.Fatalf("ExceedsRunSpan(%d, %d) = %v, want %v", tc.start, tc.to, got, tc.exceedsSpan)
			}
		})
	}
}
