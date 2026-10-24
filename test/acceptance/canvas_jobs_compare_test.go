package acceptance

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func completedJob(t *testing.T, client *bridgeClient, plan map[string]any) string {
	t.Helper()
	world := startJob(t, client, plan)
	waitForJobState(t, client, world, "completed")
	waitForListing(t, client, world, "the job "+world+" to be listed complete", func(entry map[string]any) bool { return entry["complete"] == true })
	return world
}

func TestTwoJobsCompareOverTheBridgeAndARunThatIsStillOpenIsRefused(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs three jobs through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	baseline := jobPlan(240, 0)
	baseline["faults"] = []any{}
	reference := completedJob(t, client, baseline)
	disturbed := completedJob(t, client, jobPlan(240, 0))
	first := client.action("run.compare", map[string]any{"a": reference, "b": disturbed})
	checkComparisonOfTheFaultedRun(t, first)
	again := client.action("run.compare", map[string]any{"a": reference, "b": disturbed})
	if one, other := encodedDocument(t, first), encodedDocument(t, again); one != other {
		t.Fatalf("two comparisons of the same runs differ:\n%s\n%s", one, other)
	}
	open := startJob(t, client, jobPlan(1800, 10))
	status, document, _ := client.call(http.MethodPost, "/api/actions/run.compare", map[string]any{"args": map[string]any{"a": reference, "b": open}})
	if status == http.StatusOK || errorCode(document) != "run_not_ended" {
		t.Fatalf("run.compare of a running job: %d %v, want run_not_ended", status, document)
	}
	client.action("job.stop", map[string]any{"world_id": open})
	waitForJobState(t, client, open, "stopped")
	checkHandoverOfAStoppedRun(t, client, open)
}

func encodedDocument(t *testing.T, document map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func checkComparisonOfTheFaultedRun(t *testing.T, comparison map[string]any) {
	t.Helper()
	plan, _ := comparison["plan"].(map[string]any)
	faults, _ := plan["faults"].(map[string]any)
	calls, _ := plan["effector_calls"].(map[string]any)
	if len(faults["only_in_b"].([]any)) != 1 || len(faults["only_in_a"].([]any)) != 0 || len(calls["identical"].([]any)) != 1 {
		t.Fatalf("plan = %v, want the fan failure only in B and the call identical", plan)
	}
	results, _ := comparison["results"].(map[string]any)
	rows, _ := results["kpis"].([]any)
	if len(rows) == 0 {
		t.Fatalf("results = %v, want KPI rows for the factory domain", results)
	}
	row, _ := rows[0].(map[string]any)
	scope, _ := row["scope"].(map[string]any)
	if row["presence"] != "both" || row["delta"] == nil || scope["kind"] != "whole_run" || scope["a_simulated_length_ns"] != float64(240e9) {
		t.Fatalf("kpi row = %v, want both sides, a delta and the whole-run scope over 240 s", row)
	}
	windowed, _ := comparison["windowed_results"].(map[string]any)
	if windowed["status"] != "not_available" || !strings.Contains(encodedDocument(t, comparison), "single_run_per_side") {
		t.Fatalf("comparison = %v, want windowed results not available and the single run limit", comparison)
	}
}

func checkHandoverOfAStoppedRun(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	asJSON := client.action("run.report", map[string]any{"run": world})
	if !hasQuestion(asJSON, "stopped_early") || !hasQuestion(asJSON, "replay_not_checked") {
		t.Fatalf("open questions = %v, want stopped_early and replay_not_checked", asJSON["open_questions"])
	}
	handover := client.action("run.report", map[string]any{"run": world, "format": "markdown", "replay": true})
	text, _ := handover["markdown"].(string)
	if handover["format"] != "markdown" || !strings.Contains(text, "`stopped_early`") || !strings.Contains(text, "## Open questions") || !strings.Contains(text, "| Replay | reproduces |") {
		t.Fatalf("handover = %v, want markdown that lists the early stop among the open questions and a reproducing replay", handover)
	}
	if strings.Contains(text, "`replay_not_checked`") {
		t.Fatalf("handover after a replay still asks whether the run reproduces:\n%s", text)
	}
}

func hasQuestion(report map[string]any, code string) bool {
	questions, _ := report["open_questions"].([]any)
	for _, raw := range questions {
		if question, _ := raw.(map[string]any); question["code"] == code {
			return true
		}
	}
	return false
}
