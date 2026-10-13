package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAPerturbationStopsApplyingExactlyAtUntilNS(t *testing.T) {
	t.Parallel()
	const from, until = int64(100e9), int64(200e9)
	cases := []struct {
		name    string
		atNS    int64
		applies bool
	}{
		{"one nanosecond before from", from - 1, false},
		{"exactly from", from, true},
		{"inside the window", (from + until) / 2, true},
		{"one nanosecond before until", until - 1, true},
		{"exactly until", until, false},
		{"after until", until + 1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			l := New("w", 1, testSpec(t))
			if _, err := l.Apply(Drop, map[string]any{"rate": 1.0}, from, until); err != nil {
				t.Fatal(err)
			}
			recs := l.Process(ev(0, "num", 1.0), c.atNS)
			if len(recs) != 1 || recs[0].Delivered == c.applies {
				t.Fatalf("at %d: applies=%v, got %+v", c.atNS, c.applies, recs)
			}
		})
	}
}

func TestAnOpenEndedPerturbationNeverStops(t *testing.T) {
	t.Parallel()
	l := New("w", 1, testSpec(t))
	if _, err := l.Apply(Drop, map[string]any{"rate": 1.0}, 100e9, 0); err != nil {
		t.Fatal(err)
	}
	if recs := l.Process(ev(0, "num", 1.0), 1<<62); len(recs) != 1 || recs[0].Delivered {
		t.Fatalf("until_ns 0 must keep the perturbation applying: %+v", recs)
	}
}

func TestBufferedRecordsAreReleasedExactlyAtUntilNS(t *testing.T) {
	t.Parallel()
	const from, until = int64(100e9), int64(200e9)
	l := New("w", 3, testSpec(t))
	if _, err := l.Apply(ProducerFlap, nil, from, until); err != nil {
		t.Fatal(err)
	}
	if recs := l.Process(ev(0, "num", 1.0), from); len(recs) != 0 {
		t.Fatalf("a record inside the window must be withheld: %+v", recs)
	}
	if flushed := l.Flush(until - 1); len(flushed) != 0 {
		t.Fatalf("one nanosecond before until_ns the buffer must stay: %+v", flushed)
	}
	flushed := l.Flush(until)
	if len(flushed) != 1 || flushed[0].Event.ObservedTime != model.FormatTime(until) {
		t.Fatalf("exactly at until_ns the buffer must be released at until_ns: %+v", flushed)
	}
	if recs := l.Process(ev(1, "num", 1.0), until); len(recs) != 1 || !recs[0].Delivered {
		t.Fatalf("a record at until_ns is outside the window: %+v", recs)
	}
}
