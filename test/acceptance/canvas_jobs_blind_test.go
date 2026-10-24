package acceptance

import (
	"net/http"
	"strings"
	"testing"
)

const canvasPerturbation = "duplicate_burst"

func blindPlan(explore bool) map[string]any {
	plan := jobPlan(7200, 600)
	plan["explore"] = explore
	plan["perturbations"] = []any{map[string]any{"kind": canvasPerturbation, "params": map[string]any{"rate": 0.5}, "at_s": 100}}
	return plan
}

func waitForPlannedFault(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	pollUntil(t, "the planned fault to have been injected", func() bool {
		status := jobState(t, client, world)
		return status["reached_ns"].(float64)-status["start_ns"].(float64) > 1900e9
	})
}

func pollWhileActive(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	args := map[string]any{"world_id": world}
	client.action("job.pause", args)
	for range 3 {
		jobState(t, client, world)
		client.action("job.list", nil)
		listedRun(t, client, world)
		client.action("job.step", args)
		client.action("job.rate", map[string]any{"world_id": world, "rate": 1})
		client.action("job.invoke_effector", map[string]any{"world_id": world, "entity_id": "CT-1", "effector": canvasEffector,
			"args": map[string]any{"tower_id": "CT-1", "speed": 0.5}})
	}
	client.action("job.resume", args)
	jobState(t, client, world)
}

func TestAnEvaluateJobsPlanNeverReachesTheBrowserWhileItRunsButTheEndedJobAndItsReportShowIt(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and polls an evaluate job through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	world := startJob(t, client, blindPlan(false))
	stream := client.follow("/api/runs/"+world+"/events", "")
	waitForPlannedFault(t, client, world)
	pollWhileActive(t, client, world)
	status, document, _ := client.call(http.MethodPost, "/api/actions/job.inject_fault", map[string]any{"args": map[string]any{"world_id": world, "entity_id": "CT-1", "fault": canvasFault}})
	if status == http.StatusOK || errorCode(document) != "evaluate_blind" {
		t.Fatalf("job.inject_fault on an evaluate job: %d %v, want evaluate_blind", status, document)
	}
	if active := jobState(t, client, world); active["state"] == "completed" {
		t.Fatal("the job ended before its active responses were checked")
	}
	heardActive := client.heard()
	client.action("job.stop", map[string]any{"world_id": world})
	for _, canary := range []string{canvasFault, canvasPerturbation, "fault_id"} {
		if strings.Contains(heardActive, canary) {
			i := strings.Index(heardActive, canary)
			t.Fatalf("a response while the evaluate job was active carries %q: ...%s...", canary, heardActive[max(0, i-200):min(len(heardActive), i+100)])
		}
	}
	checkObserverStreamHidesTruth(t, stream)
	checkEndedJobShowsItsPlan(t, client, world)
}

func checkEndedJobShowsItsPlan(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	ended := waitForJobState(t, client, world, "stopped")
	faults, _ := ended["faults"].([]any)
	perturbations, _ := ended["perturbations"].([]any)
	if len(faults) != 1 || len(perturbations) != 1 || faults[0].(map[string]any)["fault_id"] != "f-0" {
		t.Fatalf("the ended job says %v, want its plan with the id the world gave the fault", ended)
	}
	truth, _ := client.action("run.report", map[string]any{"run": world})["truth"].(map[string]any)
	if reported, _ := truth["faults"].([]any); len(reported) != 1 {
		t.Fatalf("the report says %v, want the injected fault", truth)
	}
}

func TestAnExploreJobStillShowsItsPlanWhileItRuns(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and polls an explore job through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	world := startJob(t, client, blindPlan(true))
	waitForPlannedFault(t, client, world)
	client.action("job.pause", map[string]any{"world_id": world})
	paused := jobState(t, client, world)
	faults, _ := paused["faults"].([]any)
	perturbations, _ := paused["perturbations"].([]any)
	if len(faults) != 1 || len(perturbations) != 1 || faults[0].(map[string]any)["fault_id"] != "f-0" {
		t.Fatalf("an active explore job says %v, want its plan", paused)
	}
	client.action("job.stop", map[string]any{"world_id": world})
	waitForJobState(t, client, world, "stopped")
}
