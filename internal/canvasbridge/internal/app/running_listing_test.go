package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func TestTheListingReportsRunningForEveryRunTheBridgeKnowsIsRunning(t *testing.T) {
	t.Parallel()
	recordedSession := openedLine + `{"type":"session","stream":"observer","mode":"connected","plan":` + sessionCheck + "}\n" + closedLine
	cases := []struct {
		name                   string
		journal                string
		created                bool
		then                   string
		running, ended, linked bool
	}{
		{name: "a created world whose journal is open", journal: openedLine, created: true, running: true},
		{name: "a created world whose journal closed but whose run was not ended", journal: openedLine + closedLine, created: true, running: true},
		{name: "a created world after run.end", journal: openedLine + closedLine, created: true, then: "run.end", ended: true},
		{name: "a created world destroyed before its run ended", journal: openedLine, created: true, then: "world.destroy"},
		{name: "a recorded run the bridge never started", journal: openedLine + closedLine, ended: true},
		{name: "a recorded session the bridge never started", journal: recordedSession, ended: true, linked: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.prepare = func(d *fakeDirector) {
				d.results["sim.world.create"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1"}`)}
			}
			h.write(t, "runs/w-1/journal.jsonl", tc.journal)
			if tc.created {
				act(t, h, "world.create", map[string]any{"domain": "plant"})
			}
			if tc.then != "" {
				act(t, h, tc.then, map[string]any{"world_id": "w-1"})
			}
			listing, err := h.bridge.Workspace()
			if err != nil || len(listing.Runs) != 1 {
				t.Fatalf("listing = %+v, %v", listing, err)
			}
			run := listing.Runs[0]
			if run.Running != tc.running || run.Ended != tc.ended || run.Connected != tc.linked {
				t.Fatalf("run = %+v, want running %t ended %t connected %t", run, tc.running, tc.ended, tc.linked)
			}
		})
	}
}

func act(t *testing.T, h *harness, action string, args map[string]any) {
	t.Helper()
	if _, err := h.bridge.RunAction(context.Background(), action, args); err != nil {
		t.Fatalf("%s: %v", action, err)
	}
}
