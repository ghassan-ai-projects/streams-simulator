package acceptance

import (
	"net/http"
	"testing"
)

func TestAPausedJobStandsStillAndTakesStepsRateChangesAndImmediateCallsThatReplay(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and operates a paced job through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	world := startJob(t, client, jobPlan(1800, 10))
	args := map[string]any{"world_id": world}
	if status, document, _ := client.call(http.MethodPost, "/api/actions/job.step", map[string]any{"args": args}); status == http.StatusOK || errorCode(document) != "job_not_paused" {
		t.Fatalf("job.step on a running job: %d %v", status, document)
	}
	paused := client.action("job.pause", args)
	first, second := jobState(t, client, world), jobState(t, client, world)
	if paused["state"] != "paused" || first["reached_ns"] != paused["reached_ns"] || second["reached_ns"] != paused["reached_ns"] || second["state"] != "paused" {
		t.Fatalf("pause = %v, then %v and %v, want the simulated time frozen across polls", paused, first, second)
	}
	if entry := listedRun(t, client, world); entry["running"] != true || entry["ended"] == true {
		t.Fatalf("a paused job listed as %v", entry)
	}
	stepped := client.action("job.step", args)
	if stepped["state"] != "paused" || stepped["reached_ns"].(float64) <= paused["reached_ns"].(float64) {
		t.Fatalf("job.step = %v, want one chunk reached and the job paused again", stepped)
	}
	checkResumedRateAndImmediateCall(t, client, world)
	status := waitForJobState(t, client, world, "stopped")
	verified := client.action("run.verify", map[string]any{"run": world})
	if verified["matches"] != true || verified["version_match"] != true {
		t.Fatalf("verify of the controlled run = %v", verified)
	}
	checkReportListsTheImmediateCall(t, client, world, status)
}

func checkResumedRateAndImmediateCall(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	args := map[string]any{"world_id": world}
	if resumed := client.action("job.resume", args); resumed["state"] != "running" {
		t.Fatalf("job.resume = %v", resumed)
	}
	if rated := client.action("job.rate", map[string]any{"world_id": world, "rate": 20}); rated["rate"] != float64(20) || rated["state"] != "running" {
		t.Fatalf("job.rate = %v", rated)
	}
	call := client.action("job.invoke_effector", map[string]any{"world_id": world, "entity_id": "CT-1", "effector": canvasEffector,
		"args": map[string]any{"tower_id": "CT-1", "speed": 0.6}})
	outcome, _ := call["call"].(map[string]any)
	if outcome["accepted"] != true || outcome["immediate"] != true || call["token"] != nil {
		t.Fatalf("job.invoke_effector = %v, want an accepted immediate outcome", call)
	}
	if stopped := client.action("job.stop", args); stopped["world_id"] != world {
		t.Fatalf("job.stop = %v", stopped)
	}
}

func checkReportListsTheImmediateCall(t *testing.T, client *bridgeClient, world string, status map[string]any) {
	t.Helper()
	calls, _ := status["calls"].([]any)
	if len(calls) != 1 || calls[0].(map[string]any)["immediate"] != true {
		t.Fatalf("the stopped job's calls = %v, want the one immediate call", status["calls"])
	}
	report := client.action("run.report", map[string]any{"run": world})
	truth, _ := report["truth"].(map[string]any)
	logged, _ := truth["effector_calls"].([]any)
	if len(logged) != 1 || logged[0].(map[string]any)["command_id"] != "job-now-1" || logged[0].(map[string]any)["effector"] != canvasEffector {
		t.Fatalf("the report lists the effector calls %v, want the immediate call job-now-1", truth["effector_calls"])
	}
	run := reportRun(t, client, world)
	if termination, _ := run["termination"].(map[string]any); termination["reason"] != "stopped_by_operator" || run["complete"] != true {
		t.Fatalf("the report says %v", run)
	}
}
