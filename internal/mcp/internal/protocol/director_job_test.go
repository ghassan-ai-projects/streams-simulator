package protocol

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func jobArguments(rate float64) map[string]any {
	return map[string]any{"domain": "aquaculture-pond", "seed": 4, "explore": true, "duration_s": 1800, "rate": rate,
		"faults": []any{map[string]any{"entity_id": "site-a/pond-1", "fault": "aerator_failure", "at_s": 300}},
		"calls":  []any{map[string]any{"entity_id": "site-a/pond-1", "effector": "start_aerator", "args": map[string]any{"pond_id": "site-a/pond-1"}, "at_s": 900}}}
}

func settledJob(t *testing.T, cs *mcp.ClientSession, worldID string) map[string]any {
	t.Helper()
	deadline := time.After(time.Minute)
	for {
		status := mustCall(t, cs, "sim.job.status", map[string]any{"world_id": worldID})
		if state := status["state"]; state != "running" && state != "stopping" {
			return status
		}
		select {
		case <-deadline:
			t.Fatalf("job %s never finished: %v", worldID, status)
		default:
			runtime.Gosched()
		}
	}
}

func TestJobToolsStartReportAndListAJobThatRunsToCompletion(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", jobArguments(0))
	worldID, _ := started["world_id"].(string)
	if worldID == "" || started["state"] != "running" || started["requested_end_ns"] == nil || started["explore"] != true {
		t.Fatalf("job.start = %v, want a running job on a world", started)
	}
	status := settledJob(t, cs, worldID)
	calls, _ := status["calls"].([]any)
	if status["state"] != "completed" || status["reached_ns"] != status["requested_end_ns"] || len(calls) != 1 {
		t.Fatalf("job.status = %v, want completed with the one call recorded", status)
	}
	listed := mustCall(t, cs, "sim.job.list", map[string]any{})
	if jobs, _ := listed["jobs"].([]any); len(jobs) != 1 {
		t.Fatalf("job.list = %v, want the finished job", listed)
	}
}

func TestJobToolsStopAPacedJobAndRefuseASecond(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", jobArguments(1))
	worldID, _ := started["world_id"].(string)
	if _, refusal := toolCall(t, cs, "sim.job.start", jobArguments(1)); !strings.HasPrefix(refusal, "job_active: ") {
		t.Fatalf("second job.start = %q, want a job_active refusal", refusal)
	}
	if stopped := mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": worldID}); stopped["state"] != "stopping" && stopped["state"] != "stopped" {
		t.Fatalf("job.stop = %v", stopped)
	}
	if status := settledJob(t, cs, worldID); status["state"] != "stopped" {
		t.Fatalf("job.status = %v, want stopped", status)
	}
}

func TestJobToolsRefuseWhatCannotRunWithTheDirectorsOwnWords(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	bad := jobArguments(0)
	bad["faults"] = []any{map[string]any{"entity_id": "site-a/pond-1", "fault": "no-such-fault", "at_s": 300}}
	if _, refusal := toolCall(t, cs, "sim.job.start", bad); !strings.Contains(refusal, "no-such-fault") {
		t.Fatalf("job.start with an impossible plan = %q, want the scenario check's refusal", refusal)
	}
	for name, change := range map[string]func(map[string]any){
		"no duration":       func(a map[string]any) { delete(a, "duration_s") },
		"duration zero":     func(a map[string]any) { a["duration_s"] = 0 },
		"negative rate":     func(a map[string]any) { a["rate"] = -1 },
		"unknown property":  func(a map[string]any) { a["pace"] = "watch" },
		"call without args": func(a map[string]any) { a["calls"] = []any{map[string]any{"entity_id": "x", "at_s": 1}} },
		"fault time as text": func(a map[string]any) {
			a["faults"] = []any{map[string]any{"entity_id": "x", "fault": "y", "at_s": "soon"}}
		},
	} {
		args := jobArguments(0)
		change(args)
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "sim.job.start", Arguments: args})
		if err == nil && !res.IsError {
			t.Errorf("%s: job.start was accepted", name)
		}
	}
	if _, refusal := toolCall(t, cs, "sim.job.status", map[string]any{"world_id": "w-9"}); !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("job.status of an unknown world = %q", refusal)
	}
	if _, refusal := toolCall(t, cs, "sim.job.stop", map[string]any{"world_id": "w-9"}); !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("job.stop of an unknown world = %q", refusal)
	}
}

func TestJobToolsAreOnTheDirectorSurfaceOnly(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	directorTools, operator := toolNames(t, NewDirectorServer(d)), toolNames(t, NewOperatorServerResolver(d))
	for _, name := range []string{"sim.job.start", "sim.job.status", "sim.job.stop", "sim.job.list"} {
		if !directorTools[name] || operator[name] {
			t.Errorf("%s: director %v operator %v, want the director only", name, directorTools[name], operator[name])
		}
	}
}
