package domain

import (
	"encoding/json"
	"testing"
)

func TestARunIsSummarisedFromItsJournalOpeningAndClosing(t *testing.T) {
	t.Parallel()
	opened := []byte(`{"type":"run.opened","mode":"explore","start_time_ns":5,"domain":{"id":"factory"}}`)
	closed := []byte("{\"type\":\"observation\"}\n{\"type\":\"run.closed\",\"at_ns\":65}\n")
	abandoned := []byte("{\"type\":\"run.closed\",\"at_ns\":65,\"incomplete\":true}\n")
	cases := map[string]struct {
		first, tail []byte
		want        RunSummary
	}{
		"closed":            {opened, closed, RunSummary{Plant: "factory", Explore: true, StartNs: 5, Closed: true, DurationNs: 60}},
		"closed incomplete": {opened, abandoned, RunSummary{Plant: "factory", Explore: true, StartNs: 5, Closed: true, Incomplete: true, DurationNs: 60}},
		"still open":        {opened, []byte(`{"type":"observation"}`), RunSummary{Plant: "factory", Explore: true, StartNs: 5}},
		"not a run":         {[]byte(`{"type":"observation"}`), closed, RunSummary{}},
		"unreadable":        {[]byte(`{`), closed, RunSummary{}},
		"empty journal":     {nil, nil, RunSummary{}},
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

func TestARunArtifactCarriesTheSeedAndCountsItsFaultCommands(t *testing.T) {
	t.Parallel()
	artifact := ParseRunArtifact([]byte(`{"seed":3,"command_log":[{"op":"fault.inject"},{"op":"clock.advance"},{"op":"fault.inject"},{"op":"fault.clear"}]}`))
	if artifact.Seed == nil || *artifact.Seed != 3 || artifact.Faults() != 2 {
		t.Fatalf("ParseRunArtifact = seed %v, faults %d", artifact.Seed, artifact.Faults())
	}
	unreadable := ParseRunArtifact([]byte("{"))
	if unreadable.Seed != nil || unreadable.Faults() != 0 {
		t.Fatalf("unreadable run.json = seed %v, faults %d", unreadable.Seed, unreadable.Faults())
	}
}

func TestANewRunEntryIsCompleteOnlyWhenItEndedWithoutBeingAbandoned(t *testing.T) {
	t.Parallel()
	seed := uint64(3)
	abandoned := RunArtifact{Seed: &seed, Error: "the world was destroyed before its run ended"}
	cases := map[string]struct {
		summary  RunSummary
		running  bool
		artifact RunArtifact
		want     RunEntry
	}{
		"finished": {RunSummary{Closed: true, DurationNs: 60}, false, RunArtifact{Seed: &seed},
			RunEntry{WorldID: "w", Ended: true, Complete: true, Seed: &seed, DurationNs: 60}},
		"abandoned": {RunSummary{Closed: true, Incomplete: true, DurationNs: 60}, false, abandoned,
			RunEntry{WorldID: "w", Ended: true, IncompleteReason: abandoned.Error, Seed: &seed, DurationNs: 60}},
		"still running": {RunSummary{Explore: true}, true, RunArtifact{},
			RunEntry{WorldID: "w", Explore: true}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := NewRunEntry("w", c.summary, c.running, c.artifact)
			if got.WorldID != c.want.WorldID || got.Ended != c.want.Ended || got.Complete != c.want.Complete ||
				got.IncompleteReason != c.want.IncompleteReason || got.Explore != c.want.Explore || got.DurationNs != c.want.DurationNs ||
				(got.Seed == nil) != (c.want.Seed == nil) || (got.Seed != nil && *got.Seed != *c.want.Seed) {
				t.Fatalf("NewRunEntry = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestARunArtifactCarriesTheRequestedEndAndTheTerminationReason(t *testing.T) {
	t.Parallel()
	stopped := ParseRunArtifact([]byte(`{"plan":{"requested_end_ns":1800000000000},"termination":{"reason":"stopped_by_operator","at_ns":5}}`))
	if stopped.RequestedEndNs() != 1800000000000 || stopped.TerminationReason() != "stopped_by_operator" {
		t.Fatalf("stopped run.json = %d, %q", stopped.RequestedEndNs(), stopped.TerminationReason())
	}
	older := ParseRunArtifact([]byte(`{"seed":1}`))
	if older.RequestedEndNs() != 0 || older.TerminationReason() != "" {
		t.Fatalf("older run.json = %d, %q", older.RequestedEndNs(), older.TerminationReason())
	}
}

func TestARunEntryStatesTheRequestedLengthFromTheRunsStartAndOnlyWhenTheRunKeptIt(t *testing.T) {
	t.Parallel()
	artifact := ParseRunArtifact([]byte(`{"plan":{"requested_end_ns":1800000000005},"termination":{"reason":"stopped_by_operator"}}`))
	stopped := wireOf(t, NewRunEntry("w", RunSummary{Closed: true, StartNs: 5, DurationNs: 540000000000}, false, artifact))
	if stopped["requested_duration_ns"] != 1.8e12 || stopped["termination_reason"] != "stopped_by_operator" {
		t.Fatalf("stopped entry = %v", stopped)
	}
	plain := wireOf(t, NewRunEntry("w", RunSummary{Closed: true}, false, RunArtifact{}))
	for _, field := range []string{"requested_duration_ns", "termination_reason"} {
		if _, has := plain[field]; has {
			t.Fatalf("plain entry = %v", plain)
		}
	}
}

func wireOf(t *testing.T, entry RunEntry) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}
