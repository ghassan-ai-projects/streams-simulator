package app

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	namedJobRunning   = `{"world_id":"w-1","run_id":"r-1","state":"running","name":"Loss and recovery","domain":"pond","explore":true,"calls":[]}`
	namedJobCompleted = `{"world_id":"w-1","run_id":"r-1","state":"completed","name":"Loss and recovery","domain":"pond","explore":true,"calls":[]}`
	unreadableJournal = "{}\n"
)

func TestTheListingOfARunningJobTakesItsNamePlantAndModeFromTheJob(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.job.start"] = process.ToolResult{Structured: []byte(namedJobRunning)}
	}
	act(t, h, "job.start", jobPlan)
	h.write(t, "runs/w-1/journal.jsonl", unreadableJournal)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
	if run := listing.Runs[0]; run.Name != "Loss and recovery" || run.Plant != "pond" || !run.Explore || !run.Running || run.Ended {
		t.Fatalf("run = %+v, want the running job's name, plant and mode", run)
	}
}

func TestOnceTheJobEndedTheListingReadsTheRecordedRunAsBefore(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.job.start"] = process.ToolResult{Structured: []byte(namedJobRunning)}
	}
	act(t, h, "job.start", jobPlan)
	h.directors[0].results["sim.job.list"] = process.ToolResult{Structured: []byte(`{"jobs":[` + namedJobCompleted + `]}`)}
	h.write(t, "runs/w-1/journal.jsonl", `{"type":"run.opened","mode":"evaluate","domain":{"id":"recorded-plant"}}`+"\n"+closedLine)
	h.write(t, "runs/w-1/run.json", `{"plan":{"name":"Recorded name"}}`)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
	if run := listing.Runs[0]; run.Name != "Recorded name" || run.Plant != "recorded-plant" || run.Explore || run.Running || !run.Ended {
		t.Fatalf("run = %+v, want everything from the recorded files once the job ended", run)
	}
}

func TestAWorldThatIsNoJobNeverGetsAJobsValues(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.world.create"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1","name":"not a job","domain":"pond","explore":true}`)}
	}
	act(t, h, "world.create", map[string]any{"domain": "plant"})
	h.write(t, "runs/w-1/journal.jsonl", unreadableJournal)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
	if run := listing.Runs[0]; run.Name != "" || run.Plant != "" || run.Explore || !run.Running {
		t.Fatalf("run = %+v, want nothing invented for a world the bridge did not start as a job", run)
	}
}

func TestTheNameOfARunningJobIsScrubbedAtTheListingBoundary(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.job.start"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1","state":"running","name":"see ` + h.dir + `/runs/w-1","domain":"pond"}`)}
	}
	act(t, h, "job.start", jobPlan)
	h.write(t, "runs/w-1/journal.jsonl", unreadableJournal)
	listing, err := boundary{h.bridge}.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 || strings.Contains(listing.Runs[0].Name, h.dir) || listing.Runs[0].Name != "see runs/w-1" {
		t.Fatalf("listing = %+v, %v, want the workspace folder scrubbed from the name", listing.Runs, err)
	}
}
