package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	jobRunning   = `{"world_id":"w-1","run_id":"r-1","state":"running","calls":[]}`
	jobPaused    = `{"world_id":"w-1","run_id":"r-1","state":"paused","calls":[]}`
	jobStopping  = `{"world_id":"w-1","run_id":"r-1","state":"stopping","calls":[]}`
	jobCompleted = `{"world_id":"w-1","run_id":"r-1","state":"completed","calls":[]}`
	jobFailed    = `{"world_id":"w-1","run_id":"r-1","state":"failed","failure":"job failed","calls":[]}`
	jobStopped   = `{"world_id":"w-1","run_id":"r-1","state":"stopped","calls":[]}`
)

var jobPlan = map[string]any{"domain": "plant", "seed": 3.0, "duration_s": 60.0, "faults": []any{}, "calls": []any{}}

func jobHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.job.start"] = process.ToolResult{Structured: []byte(jobRunning)}
	}
	return h
}

func jobArguments(action string) map[string]any {
	switch action {
	case "job.list":
		return nil
	case "job.rate":
		return map[string]any{"world_id": "w-1", "rate": 2.0}
	case "job.inject_fault":
		return map[string]any{"world_id": "w-1", "entity_id": "e", "fault": "f"}
	case "job.invoke_effector":
		return map[string]any{"world_id": "w-1", "entity_id": "e", "effector": "x"}
	}
	return map[string]any{"world_id": "w-1"}
}

func startJob(t *testing.T, h *harness) {
	t.Helper()
	act(t, h, "job.start", jobPlan)
	h.write(t, "runs/w-1/journal.jsonl", openedLine)
}

func TestStartingAJobTracksItsWorldLikeACreatedWorld(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	if !h.bridge.runLive() || !h.bridge.runIsRunning("w-1") {
		t.Fatal("the world of a started job is not tracked as running")
	}
	h.write(t, plant, "{}")
	act(t, h, "catalog.list", nil)
	if len(h.directors) != 1 || h.directors[0].closed {
		t.Fatalf("the director restarted under a live job: %d directors", len(h.directors))
	}
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 || !listing.Runs[0].Running || listing.Runs[0].Ended {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
}

func TestEveryJobResultSettlesTheRunningFlagOfItsWorld(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, action, result string
		running              bool
	}{
		{"status of a running job", "job.status", jobRunning, true},
		{"status of a stopping job", "job.status", jobStopping, true},
		{"status of a paused job", "job.status", jobPaused, true},
		{"pause of a running job", "job.pause", jobPaused, true},
		{"resume of a paused job", "job.resume", jobRunning, true},
		{"step of a paused job", "job.step", jobPaused, true},
		{"a rate change", "job.rate", jobRunning, true},
		{"an immediate fault", "job.inject_fault", `{"world_id":"w-1","state":"paused","fault_id":"f-1","calls":[]}`, true},
		{"an immediate call on a job that had just finished", "job.invoke_effector", jobCompleted, false},
		{"a pause that found the job finished", "job.pause", jobCompleted, false},
		{"status of a completed job", "job.status", jobCompleted, false},
		{"status of a stopped job", "job.status", jobStopped, false},
		{"status of a failed job", "job.status", jobFailed, false},
		{"stop that is still stopping", "job.stop", jobStopping, true},
		{"stop of a job that had finished", "job.stop", jobCompleted, false},
		{"a list holding the finished job", "job.list", `{"jobs":[` + jobCompleted + `]}`, false},
		{"a list holding the running job", "job.list", `{"jobs":[` + jobRunning + `]}`, true},
		{"an empty list", "job.list", `{"jobs":[]}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := jobHarness(t)
			startJob(t, h)
			h.directors[0].results["sim."+tc.action] = process.ToolResult{Structured: []byte(tc.result)}
			act(t, h, tc.action, jobArguments(tc.action))
			if got := h.bridge.runIsRunning("w-1"); got != tc.running {
				t.Fatalf("running = %v after %s, want %v", got, tc.action, tc.running)
			}
		})
	}
}

func TestTheListingFlipsAJobThatFinishedWhileNoOneWasLooking(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	h.write(t, "runs/w-1/sealed/journal.jsonl", "truth\n")
	if _, err := h.bridge.ReadFile(sealedFile); !errors.Is(err, domain.ErrSealed) {
		t.Fatalf("the sealed journal of a running job: %v", err)
	}
	h.directors[0].results["sim.job.list"] = process.ToolResult{Structured: []byte(`{"jobs":[` + jobCompleted + `]}`)}
	h.write(t, "runs/w-1/journal.jsonl", openedLine+closedLine)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 || listing.Runs[0].Running || !listing.Runs[0].Ended {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
	if document, err := h.bridge.ReadFile(sealedFile); err != nil || document.Content != "truth\n" {
		t.Fatalf("the sealed journal after the job ended: %+v, %v", document, err)
	}
	if h.bridge.runLive() {
		t.Fatal("a finished job still keeps the director alive")
	}
}

func TestTheListingKeepsALiveJobRunningAndSealed(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	h.directors[0].results["sim.job.list"] = process.ToolResult{Structured: []byte(`{"jobs":[` + jobStopping + `]}`)}
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 || !listing.Runs[0].Running || listing.Runs[0].Ended {
		t.Fatalf("listing = %+v, %v", listing.Runs, err)
	}
	if !slices.Contains(h.directors[0].calls, "sim.job.list") {
		t.Fatalf("calls = %v, want the listing to ask the director for its jobs", h.directors[0].calls)
	}
}

func TestReconcilingNeverStartsADirector(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		setup func(*harness)
	}{
		{"nothing runs", func(*harness) {}},
		{"a tracked world whose director is gone", func(h *harness) {
			h.bridge.running["w-1"] = true
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			tc.setup(h)
			h.write(t, "runs/w-1/journal.jsonl", openedLine)
			if _, err := h.bridge.Workspace(context.Background()); err != nil || len(h.directors) != 0 {
				t.Fatalf("Workspace() started %d directors, err %v", len(h.directors), err)
			}
			if h.bridge.runLive() {
				t.Fatal("worlds of a director that is gone are still counted as open")
			}
		})
	}
}

func TestAnIdleListingDoesNotAskTheDirectorForJobs(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	act(t, h, "catalog.list", nil)
	if _, err := h.bridge.Workspace(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(h.directors[0].calls, []string{"sim.catalog.list"}) {
		t.Fatalf("calls = %v, want no job.list while nothing is running", h.directors[0].calls)
	}
}

func TestADirectorThatDiedBeforeTheListingLosesItsJobsAndIsNotReplaced(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	h.directors[0].failWith = fmt.Errorf("%w: sim.job.list: broken pipe", process.ErrGone)
	listing, err := h.bridge.Workspace(context.Background())
	if err != nil || len(listing.Runs) != 1 || listing.Runs[0].Running || listing.Runs[0].Ended {
		t.Fatalf("listing = %+v, %v, want an interrupted run: not running and not ended", listing.Runs, err)
	}
	if len(h.directors) != 1 || !h.directors[0].closed || h.bridge.runLive() {
		t.Fatalf("%d directors, closed %v: the dead director must be dropped, not replaced", len(h.directors), h.directors[0].closed)
	}
}

func TestABusyDirectorLeavesTheListingAsItWas(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	calls := len(h.directors[0].calls)
	h.bridge.directorMu.Lock()
	listing, err := h.bridge.Workspace(context.Background())
	h.bridge.directorMu.Unlock()
	if err != nil || len(listing.Runs) != 1 || !listing.Runs[0].Running || len(h.directors[0].calls) != calls {
		t.Fatalf("listing = %+v, %v after %d calls", listing.Runs, err, len(h.directors[0].calls))
	}
}

func TestAJobStartLostWithItsDirectorIsRetriedOnlyAfterTheDeadDirectorIsClosed(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	h.directors[0].failWith = fmt.Errorf("%w: sim.job.start: broken pipe", process.ErrGone)
	h.prepare = func(d *fakeDirector) {
		d.results["sim.job.start"] = process.ToolResult{Structured: []byte(`{"world_id":"w-2","state":"running"}`)}
	}
	act(t, h, "job.start", jobPlan)
	if len(h.directors) != 2 || !h.directors[0].closed || !slices.Equal(h.directors[1].calls, []string{"sim.job.start"}) {
		t.Fatalf("%d directors: the dead director must be closed and the start sent once to the new one", len(h.directors))
	}
	if h.bridge.runIsRunning("w-1") || !h.bridge.runIsRunning("w-2") {
		t.Fatal("the job of the dead director is still listed as running, or the new job is not")
	}
}

func TestJobResultsAreScrubbedAndKeepTheCapabilityOut(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	startJob(t, h)
	secret := `{"world_id":"w-1","state":"failed","failure":"open ` + h.dir + `/runs/w-1: denied","token":"t-secret","operator_endpoint":"http://127.0.0.1:1"}`
	h.directors[0].results["sim.job.start"] = process.ToolResult{Structured: []byte(secret)}
	h.directors[0].results["sim.job.status"] = process.ToolResult{Structured: []byte(secret)}
	for _, action := range []string{"job.start", "job.status"} {
		got, err := h.bridge.RunAction(context.Background(), action, jobArgumentsFor(action))
		if err != nil || strings.Contains(string(got), h.dir) || !strings.Contains(string(got), "open runs/w-1: denied") {
			t.Fatalf("%s = %s, %v", action, got, err)
		}
	}
	got, _ := h.bridge.RunAction(context.Background(), "job.start", jobPlan)
	if strings.Contains(string(got), "t-secret") || strings.Contains(string(got), "operator_endpoint") {
		t.Fatalf("job.start leaked the capability: %s", got)
	}
}

func jobArgumentsFor(action string) map[string]any {
	if action == "job.start" {
		return jobPlan
	}
	return jobArguments(action)
}

func TestAJobPlanWithAnArgumentTheRegistryDoesNotNameNeverReachesTheDirector(t *testing.T) {
	t.Parallel()
	h := jobHarness(t)
	bad := map[string]any{"domain": "plant", "duration_s": 60.0, "token": "t-x"}
	if _, err := h.bridge.RunAction(context.Background(), "job.start", bad); !errors.Is(err, domain.ErrBadArguments) || len(h.directors) != 0 {
		t.Fatalf("job.start with a token: %v after %d directors", err, len(h.directors))
	}
}
