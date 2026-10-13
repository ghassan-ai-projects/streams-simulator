package model

import (
	"slices"
	"testing"
)

func TestAKindListsEachLinkTypeItReadsOnce(t *testing.T) {
	t.Parallel()
	linked := func(state, linkType string) F1Input {
		return F1Input{State: state, Coef: 1, Link: &InputLink{Type: linkType, Aggregate: AggregateSum}}
	}
	kind := &Kind{Dynamics: []Dynamics{
		{Target: "a", Follow: &Follow{Sources: []FollowSource{{Link: "feed", State: "x"}}, Aggregate: AggregateSum}},
		{Target: "b", F1: &F1Dyn{Inputs: []F1Input{{State: "a", Coef: 1}, linked("x", "feed"), linked("y", "supply")}}},
		{Target: "c", F0: &F0Dyn{}},
	}}
	if got := kind.LinkReads(); !slices.Equal(got, []string{"feed", "supply"}) {
		t.Fatalf("link reads = %v", got)
	}
	if !(InputLink{Aggregate: AggregateMean}).NeedsDefault() || (InputLink{Aggregate: AggregateCountRunning}).NeedsDefault() {
		t.Fatal("mean, min and max need a default; sum and count_running do not")
	}
}

func TestADynamicsStepsEveryDefaultStepUnlessItDeclaresOne(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		dyn  Dynamics
		want int64
	}{
		{"none declared", Dynamics{}, DefaultStepMs},
		{"a negative step", Dynamics{DTMs: -5}, DefaultStepMs},
		{"a declared step", Dynamics{DTMs: 250}, 250},
	}
	for _, tc := range cases {
		if got := tc.dyn.StepMs(); got != tc.want {
			t.Errorf("%s: StepMs = %d, want %d", tc.name, got, tc.want)
		}
	}
}
