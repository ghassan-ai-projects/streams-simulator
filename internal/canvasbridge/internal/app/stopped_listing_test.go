package app

import (
	"context"
	"testing"
)

const startedAt = `{"type":"run.opened","mode":"evaluate","start_time_ns":100}` + "\n"

func TestTheListingCarriesTheLengthTheRunWasAskedForAndHowItTerminated(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, runJSON, reason string
		askedNs               int64
	}{
		{"a stopped run", `{"seed":7,"plan":{"requested_end_ns":1800000000100},"termination":{"reason":"stopped_by_operator","at_ns":540000000100}}`, "stopped_by_operator", 1800000000000},
		{"a completed run", `{"seed":7,"plan":{"requested_end_ns":1800000000100},"termination":{"reason":"completed","at_ns":1800000000100}}`, "completed", 1800000000000},
		{"a plan that ends before the run began", `{"plan":{"requested_end_ns":50}}`, "", 0},
		{"a run recorded before the plan was kept", `{"seed":7}`, "", 0},
		{"an unreadable run.json", `{`, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.write(t, "runs/w-1/journal.jsonl", startedAt+closedLine)
			h.write(t, "runs/w-1/run.json", tc.runJSON)
			listing, err := h.bridge.Workspace(context.Background())
			if err != nil || len(listing.Runs) != 1 {
				t.Fatalf("listing = %+v, %v", listing, err)
			}
			run := listing.Runs[0]
			if run.RequestedDurationNs != tc.askedNs || run.TerminationReason != tc.reason {
				t.Fatalf("run = %+v, want asked %d reason %q", run, tc.askedNs, tc.reason)
			}
		})
	}
}
