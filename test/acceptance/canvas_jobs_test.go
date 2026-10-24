package acceptance

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAJobRunsToCompletionAndItsEvidenceOpensWithoutAnyBrowserAction(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs a planned experiment as a job through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	world := startJob(t, client, jobPlan(120, 0))
	status := waitForJobState(t, client, world, "completed")
	calls, _ := status["calls"].([]any)
	if len(calls) != 1 || calls[0].(map[string]any)["accepted"] != true || status["reached_ns"] != status["requested_end_ns"] {
		t.Fatalf("a completed job reports %v", status)
	}
	entry := listedRun(t, client, world)
	if entry["running"] == true || entry["ended"] != true || entry["complete"] != true {
		t.Fatalf("a completed job listed as %v", entry)
	}
	checkEvidence(t, client, world)
	run := reportRun(t, client, world)
	termination, _ := run["termination"].(map[string]any)
	if termination["reason"] != "completed" || run["requested_end_ns"] != status["requested_end_ns"] || run["end_time_ns"] != status["requested_end_ns"] {
		t.Fatalf("the report says %v", run)
	}
	if token := operatorToken.FindString(client.heard()); token != "" || strings.Contains(client.heard(), "operator_endpoint") {
		t.Fatalf("a job let the operator capability through: %q", token)
	}
}

func TestAStoppedJobRecordsWhyAndHowFarItGot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and stops a paced job through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	world := startJob(t, client, jobPlan(1800, 60))
	pollUntil(t, "the job to make progress", func() bool {
		status := jobState(t, client, world)
		return status["reached_ns"].(float64) > status["start_ns"].(float64)
	})
	if stopped := client.action("job.stop", map[string]any{"world_id": world}); stopped["world_id"] != world {
		t.Fatalf("job.stop = %v", stopped)
	}
	status := waitForJobState(t, client, world, "stopped")
	client.action("job.stop", map[string]any{"world_id": world})
	if status["reached_ns"].(float64) >= status["requested_end_ns"].(float64) {
		t.Fatalf("a stopped job reached its requested end: %v", status)
	}
	waitForListing(t, client, world, "the stopped job to be listed complete", func(entry map[string]any) bool {
		return entry["running"] != true && entry["ended"] == true && entry["complete"] == true
	})
	run := reportRun(t, client, world)
	termination, _ := run["termination"].(map[string]any)
	if termination["reason"] != "stopped_by_operator" || termination["at_ns"] != status["reached_ns"] || run["end_time_ns"] != status["reached_ns"] || run["requested_end_ns"] != status["requested_end_ns"] {
		t.Fatalf("the report says %v for the job %v", run, status)
	}
	if verified := client.action("run.verify", map[string]any{"run": world}); verified["matches"] != true || verified["version_match"] != true {
		t.Fatalf("verify = %v", verified)
	}
}

func TestAReloadedBrowserFindsTheLiveJobAndTheListingSettlesItOnItsOwn(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and follows a paced job from two browsers")
	}
	client, bridge, _ := pairedJobBridge(t)
	world := startJob(t, client, jobPlan(900, 300))
	stream := client.follow("/api/runs/"+world+"/events", "")
	reloaded := reloadedBrowser(t, client, bridge)
	listed := reloaded.action("job.list", nil)
	jobs, _ := listed["jobs"].([]any)
	if len(jobs) != 1 || jobs[0].(map[string]any)["world_id"] != world || jobs[0].(map[string]any)["requested_end_ns"] == nil {
		t.Fatalf("a reloaded browser lists %v", listed)
	}
	if entry := listedRun(t, reloaded, world); entry["running"] != true || entry["ended"] == true {
		t.Fatalf("a live job listed as %v", entry)
	}
	checkSealedWhileOpen(t, reloaded, world)
	checkOneJobAtATime(t, reloaded)
	waitForListing(t, reloaded, world, "the listing to settle the finished job without a status call", func(entry map[string]any) bool {
		return entry["running"] != true
	})
	checkObserverStreamHidesTruth(t, stream)
	checkEvidence(t, reloaded, world)
	if done := jobState(t, reloaded, world); done["state"] != "completed" {
		t.Fatalf("the job after the listing settled it: %v", done)
	}
}

func checkOneJobAtATime(t *testing.T, client *bridgeClient) {
	t.Helper()
	before := listedRunCount(t, client)
	status, document, _ := client.call(http.MethodPost, "/api/actions/job.start", map[string]any{"args": jobPlan(60, 0)})
	if status == http.StatusOK || errorCode(document) != "job_active" || listedRunCount(t, client) != before {
		t.Fatalf("a second job while one runs: %d %v", status, document)
	}
}

func checkObserverStreamHidesTruth(t *testing.T, stream *eventStream) {
	t.Helper()
	events := stream.drain()
	if len(events) == 0 || events[len(events)-1].name != "closed" {
		t.Fatalf("the job's stream did not end with a closed event: %+v", events)
	}
	for _, event := range events {
		if strings.Contains(event.data, canvasFault) {
			t.Fatalf("the observer stream of a job carries sealed truth: %s", event.data)
		}
	}
}

func TestAnInvalidJobPlanIsRefusedAndLeavesNoRun(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and refuses job plans through the bridge")
	}
	client, _, workspace := pairedJobBridge(t)
	cases := []struct {
		name, code string
		spoil      func(plan map[string]any)
	}{
		{"an unknown fault", "scenario_invalid", func(plan map[string]any) { plan["faults"].([]any)[0].(map[string]any)["fault"] = "no_such_fault" }},
		{"an unknown entity", "scenario_invalid", func(plan map[string]any) { plan["calls"].([]any)[0].(map[string]any)["entity_id"] = "NOPE-9" }},
		{"a call past the end", "invalid_args", func(plan map[string]any) { plan["calls"].([]any)[0].(map[string]any)["at_s"] = 9999 }},
		{"no duration", "director_refused", func(plan map[string]any) { delete(plan, "duration_s") }},
		{"an unknown plan key", "bad_arguments", func(plan map[string]any) { plan["sink"] = "file" }},
		{"an unknown domain", "domain_invalid", func(plan map[string]any) { plan["domain"] = "no-such-domain" }},
		{"a nested unknown key", "director_refused", func(plan map[string]any) { plan["faults"].([]any)[0].(map[string]any)["token"] = "t-x" }},
	}
	for _, tc := range cases {
		plan := jobPlan(60, 0)
		tc.spoil(plan)
		status, document, _ := client.call(http.MethodPost, "/api/actions/job.start", map[string]any{"args": plan})
		if status == http.StatusOK || errorCode(document) != tc.code {
			t.Errorf("%s: %d %v, want a refusal %s", tc.name, status, document, tc.code)
		}
	}
	if count := listedRunCount(t, client); count != 0 {
		t.Errorf("refused plans left %d runs listed", count)
	}
	if entries, err := os.ReadDir(filepath.Join(workspace, "runs")); err == nil && len(entries) != 0 {
		t.Errorf("refused plans left %d run folders", len(entries))
	}
}
