package domain

import (
	"math"
	"testing"
)

func TestAnAssignmentApproachesItsTargetFromTheValueItReplaces(t *testing.T) {
	t.Parallel()
	const tau = 60e9
	e := math.Exp(-1)
	cases := []struct {
		name  string
		kicks []*kick
		at    int64
		want  float64
	}{
		{"from the natural value", []*kick{{delta: 10, assign: true, startNS: 100, tauNS: tau}}, 100 + tau, 10 - 8*e},
		{"a step assignment", []*kick{{delta: 10, assign: true, startNS: 100}}, 101, 10},
		{"from an earlier additive kick", []*kick{{delta: 3, startNS: 50}, {delta: 10, assign: true, startNS: 100, tauNS: tau}}, 100 + tau, 10 - 5*e},
		{"from an earlier assignment", []*kick{{delta: 4, assign: true, startNS: 50}, {delta: 10, assign: true, startNS: 100, tauNS: tau}}, 100 + tau, 10 - 6*e},
		{"with a later additive kick", []*kick{{delta: 10, assign: true, startNS: 100, tauNS: tau}, {delta: 1, startNS: 200}}, 100 + tau, 11 - 8*e},
		{"before it starts", []*kick{{delta: 10, assign: true, startNS: 100, tauNS: tau}}, 100, 2},
	}
	for _, tc := range cases {
		if got := drivenValue(2, tc.kicks, tc.at, tc.at); math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}
