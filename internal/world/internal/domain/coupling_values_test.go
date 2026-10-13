package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestEachAggregateOverLinkedValuesAndOverNone(t *testing.T) {
	t.Parallel()
	cases := []struct {
		form        string
		values      []float64
		want, empty float64
	}{
		{model.AggregateSum, []float64{2, 0.5, 4}, 6.5, 0},
		{model.AggregateMean, []float64{2, 0.5, 4}, 6.5 / 3, 9},
		{model.AggregateMin, []float64{2, 0.5, 4}, 0.5, 9},
		{model.AggregateMax, []float64{2, 0.5, 4}, 4, 9},
		{model.AggregateCountRunning, []float64{1, 0.49, 0.5, 0}, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.form, func(t *testing.T) {
			t.Parallel()
			agg := linkAggregate{form: tc.form}
			for _, v := range tc.values {
				agg.add(v)
			}
			if got := agg.value(9); got != tc.want {
				t.Fatalf("%s of %v = %v, want %v", tc.form, tc.values, got, tc.want)
			}
			if got := (linkAggregate{form: tc.form}).value(9); got != tc.empty {
				t.Fatalf("%s of nothing = %v, want %v", tc.form, got, tc.empty)
			}
		})
	}
}

func TestALinkedInstanceCountsFromItsBirthThroughItsRetirement(t *testing.T) {
	t.Parallel()
	alive := &Entity{BornNS: 100, alive: true}
	retired := &Entity{BornNS: 100, retiredNS: 500}
	for _, tc := range []struct {
		ent     *Entity
		at      int64
		present bool
		until   int64
	}{
		{alive, 99, false, 99}, {alive, 100, true, 100}, {alive, 900, true, 900},
		{retired, 500, true, 500}, {retired, 501, false, 500}, {retired, 300, true, 300},
	} {
		if got := tc.ent.presentAt(tc.at); got != tc.present {
			t.Fatalf("present at %d = %v, want %v", tc.at, got, tc.present)
		}
		if got := tc.ent.lastAliveAt(tc.at); got != tc.until {
			t.Fatalf("advanced until %d for %d, want %d", got, tc.at, tc.until)
		}
	}
}
