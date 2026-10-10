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

func TestReorderPreservesDeliveryOrderAndReason(t *testing.T) {
	t.Parallel()
	swaps := 0
	for seed := uint64(0); seed < 32; seed++ {
		l := New("w", seed, testSpec(t))
		if _, err := l.Apply(Reorder, map[string]any{"max_displacement": 1}, 0, 0); err != nil {
			t.Fatal(err)
		}
		if got := l.Process(ev(0, "num", 0.0), 1000000001); len(got) != 0 {
			t.Fatalf("first record should remain buffered, got %+v", got)
		}
		got := l.Process(ev(1, "num", 1.0), 1000000002)
		switch {
		case len(got) == 2 && got[0].Event.Seq == 1 && got[1].Event.Seq == 0:
			swaps++
			for _, r := range got {
				if r.Reason != model.DeliveryReordered {
					t.Fatalf("seed %d: reordered delivery lacks ledger reason: %+v", seed, r)
				}
			}
		case len(got) == 1 && got[0].Event.Seq == 0:
			// The pair was left in order by this seed's draw.
		default:
			t.Fatalf("seed %d: records were lost or duplicated before delivery: %+v", seed, got)
		}
	}
	if swaps == 0 || swaps == 32 {
		t.Fatalf("%d of 32 seeds swapped: the swap must be a random draw, not fixed", swaps)
	}
}
