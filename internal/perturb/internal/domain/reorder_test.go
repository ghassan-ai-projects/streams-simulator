package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

// Whether a pair is swapped is the perturbation's random draw, not a property
// of the delivery time: events on whole seconds must not all swap.
func TestReorderSwapsAboutHalfOfTheWholeSecondPairs(t *testing.T) {
	t.Parallel()
	window := newReorderWindow(1, randutil.NewSplitMix64(7))
	var out []Delivered
	for i := range 400 {
		at := int64(i) * 1e9
		record := Delivered{Event: model.SimEvent{Seq: int64(i)}, Delivered: true}
		out = append(out, window.push([]Delivered{record}, at)...)
	}
	out = append(out, window.flush(0)...)
	swapped := 0
	for i := 1; i < len(out); i++ {
		if out[i].Event.Seq < out[i-1].Event.Seq {
			swapped++
		}
	}
	if swapped < 90 || swapped > 180 {
		t.Fatalf("%d adjacent inversions in 400 whole-second records, want a random half of the ~265 pops", swapped)
	}
}

func TestReorderIsDeterministicForASeed(t *testing.T) {
	t.Parallel()
	order := func() []int64 {
		window := newReorderWindow(1, randutil.NewSplitMix64(3))
		var seqs []int64
		for i := range 50 {
			for _, r := range window.push([]Delivered{{Event: model.SimEvent{Seq: int64(i)}}}, int64(i)) {
				seqs = append(seqs, r.Event.Seq)
			}
		}
		return seqs
	}
	a, b := order(), order()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("reorder is not deterministic at %d: %v vs %v", i, a, b)
		}
	}
}
