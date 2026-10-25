package protocol

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJobControlToolsPauseStepRateAndResumeAPacedJob(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", jobArguments(1))
	world := map[string]any{"world_id": started["world_id"]}
	paused := mustCall(t, cs, "sim.job.pause", world)
	frozen := mustCall(t, cs, "sim.job.status", world)
	if paused["state"] != "paused" || frozen["state"] != "paused" || frozen["reached_ns"] != paused["reached_ns"] {
		t.Fatalf("pause = %v, then status %v, want a paused job whose time stands still", paused, frozen)
	}
	stepped := mustCall(t, cs, "sim.job.step", world)
	if stepped["state"] != "paused" || stepped["reached_ns"].(float64) <= paused["reached_ns"].(float64) {
		t.Fatalf("step = %v, want one more chunk reached and the job still paused", stepped)
	}
	if rated := mustCall(t, cs, "sim.job.rate", map[string]any{"world_id": started["world_id"], "rate": 0}); rated["rate"] != float64(0) {
		t.Fatalf("rate = %v, want rate 0", rated)
	}
	if resumed := mustCall(t, cs, "sim.job.resume", world); resumed["state"] != "running" && resumed["state"] != "completed" {
		t.Fatalf("resume = %v, want the job going again", resumed)
	}
	if status := settledJob(t, cs, started["world_id"].(string)); status["state"] != "completed" {
		t.Fatalf("status = %v, want the controlled job completed", status)
	}
}

func TestJobCommandToolsApplyAFaultAndAnEffectorCallNow(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", jobArguments(1))
	id := started["world_id"]
	mustCall(t, cs, "sim.job.pause", map[string]any{"world_id": id})
	fault := mustCall(t, cs, "sim.job.inject_fault", map[string]any{"world_id": id, "entity_id": "site-a/pond-1", "fault": "do_probe_fouling"})
	if fault["fault_id"] == "" || fault["fault_id"] == nil || fault["state"] != "paused" {
		t.Fatalf("inject_fault = %v, want a fault id on the paused job", fault)
	}
	call := mustCall(t, cs, "sim.job.invoke_effector", map[string]any{"world_id": id, "entity_id": "site-a/pond-1", "effector": "start_aerator",
		"args": map[string]any{"pond_id": "site-a/pond-1"}})
	outcome, _ := call["call"].(map[string]any)
	if outcome["immediate"] != true || outcome["accepted"] != true || outcome["effector"] != "start_aerator" {
		t.Fatalf("invoke_effector = %v, want an accepted immediate outcome", call)
	}
	mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": id})
	settledJob(t, cs, id.(string))
}

func TestJobControlToolsRefuseWhatCannotApplyWithTheDirectorsOwnWords(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", jobArguments(1))
	id := started["world_id"]
	for name, tc := range map[string]struct {
		tool, want string
		args       map[string]any
	}{
		"step while running":      {"sim.job.step", "job_not_paused: ", map[string]any{"world_id": id}},
		"unknown world":           {"sim.job.pause", "world_not_found: ", map[string]any{"world_id": "w-9"}},
		"unknown effector":        {"sim.job.invoke_effector", "unknown_effector: ", map[string]any{"world_id": id, "entity_id": "site-a/pond-1", "effector": "nope"}},
		"unknown fault":           {"sim.job.inject_fault", "domain_invalid: ", map[string]any{"world_id": id, "entity_id": "site-a/pond-1", "fault": "nope"}},
		"rate below zero":         {"sim.job.rate", "", map[string]any{"world_id": id, "rate": -1}},
		"rate missing":            {"sim.job.rate", "", map[string]any{"world_id": id}},
		"unknown property":        {"sim.job.pause", "", map[string]any{"world_id": id, "at": 3}},
		"fault without entity":    {"sim.job.inject_fault", "", map[string]any{"world_id": id, "fault": "x"}},
		"effector without entity": {"sim.job.invoke_effector", "", map[string]any{"world_id": id, "effector": "x"}},
	} {
		if tc.want != "" {
			if _, refusal := toolCall(t, cs, tc.tool, tc.args); !strings.HasPrefix(refusal, tc.want) {
				t.Errorf("%s: refusal = %q, want prefix %q", name, refusal, tc.want)
			}
			continue
		}
		if res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: tc.tool, Arguments: tc.args}); err == nil && !res.IsError {
			t.Errorf("%s: was accepted", name)
		}
	}
	mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": id})
	settledJob(t, cs, id.(string))
}
