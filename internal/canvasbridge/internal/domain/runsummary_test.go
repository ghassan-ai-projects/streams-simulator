package domain

import "testing"

func TestARunIsSummarisedFromItsJournalOpeningAndClosing(t *testing.T) {
	t.Parallel()
	opened := []byte(`{"type":"run.opened","mode":"explore","start_time_ns":5,"domain":{"id":"factory"}}`)
	closed := []byte("{\"type\":\"observation\"}\n{\"type\":\"run.closed\",\"at_ns\":65}\n")
	cases := map[string]struct {
		first, tail []byte
		want        RunSummary
	}{
		"closed":        {opened, closed, RunSummary{Plant: "factory", Explore: true, Closed: true, DurationNs: 60}},
		"still open":    {opened, []byte(`{"type":"observation"}`), RunSummary{Plant: "factory", Explore: true}},
		"not a run":     {[]byte(`{"type":"observation"}`), closed, RunSummary{}},
		"unreadable":    {[]byte(`{`), closed, RunSummary{}},
		"empty journal": {nil, nil, RunSummary{}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := SummarizeRun(c.first, c.tail); got != c.want {
				t.Fatalf("SummarizeRun = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestARunEndsOnlyWhenClosedAndNoLongerRunning(t *testing.T) {
	t.Parallel()
	closed := RunSummary{Closed: true, Explore: true}
	if got := closed.State(false); got != (RunState{Ended: true, Explore: true}) {
		t.Fatalf("closed, not running = %+v", got)
	}
	if got := closed.State(true); got.Ended {
		t.Fatalf("closed but running = %+v", got)
	}
	if got := (RunSummary{}).State(false); got != (RunState{}) {
		t.Fatalf("unreadable journal = %+v", got)
	}
}
