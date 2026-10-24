package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestARateChangeIsCheckedLikeTheRequestsRate(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		rate float64
		ok   bool
	}{
		"as fast as possible": {0, true}, "a positive rate": {2.5, true},
		"negative": {-1, false}, "not a number": {math.NaN(), false}, "infinite": {math.Inf(1), false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := CheckJobRate(tc.rate)
			var tool *ToolError
			if tc.ok != (err == nil) || (!tc.ok && (!errors.As(err, &tool) || tool.Code != CodeInvalidArgs || !strings.Contains(tool.Msg, "rate"))) {
				t.Fatalf("CheckJobRate(%v) = %v, want ok=%v", tc.rate, err, tc.ok)
			}
		})
	}
}

func TestAnImmediateCommandNeedsItsEntityAndItsName(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		err  error
		want string
	}{
		"fault without entity":     {CheckImmediateFault("", "f"), "entity_id"},
		"fault without name":       {CheckImmediateFault("e", ""), "fault"},
		"call without entity":      {CheckImmediateCall("", "f"), "entity_id"},
		"call without an effector": {CheckImmediateCall("e", ""), "effector"},
	}
	for name, tc := range cases {
		var tool *ToolError
		if !errors.As(tc.err, &tool) || tool.Code != CodeInvalidArgs || !strings.Contains(tool.Msg, tc.want) {
			t.Errorf("%s: err = %v, want an invalid_args refusal naming %q", name, tc.err, tc.want)
		}
	}
	if CheckImmediateFault("e", "f") != nil || CheckImmediateCall("e", "f") != nil {
		t.Fatal("a complete immediate command was refused")
	}
}

func TestImmediateCommandIdsAreNumberedBesideThePlannedOnes(t *testing.T) {
	t.Parallel()
	if got := ImmediateCommandID(3); got != "job-now-3" {
		t.Fatalf("ImmediateCommandID(3) = %q, want job-now-3", got)
	}
}
