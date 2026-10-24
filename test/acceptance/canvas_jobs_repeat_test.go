package acceptance

import (
	"math"
	"testing"
)

func TestARunSetUpAgainFromItsReportReproducesTheSameTrace(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs two jobs through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	first := startJob(t, client, jobPlan(240, 0))
	waitForJobState(t, client, first, "completed")
	waitForListing(t, client, first, "the first job to be listed complete", func(entry map[string]any) bool { return entry["complete"] == true })
	again := startJob(t, client, planFromReport(t, client.action("run.report", map[string]any{"run": first})))
	waitForJobState(t, client, again, "completed")
	waitForListing(t, client, again, "the repeated job to be listed complete", func(entry map[string]any) bool { return entry["complete"] == true })
	original, repeated := client.action("run.report", map[string]any{"run": first}), client.action("run.report", map[string]any{"run": again})
	for _, digest := range []string{"trace", "world", "ledger"} {
		if got, want := digestOf(t, repeated, digest), digestOf(t, original, digest); got != want {
			t.Errorf("the repeated run's %s digest is %s, the original's is %s", digest, got, want)
		}
	}
	if verified := client.action("run.verify", map[string]any{"run": again}); verified["matches"] != true {
		t.Fatalf("verify of the repeated run = %v", verified)
	}
}

func digestOf(t *testing.T, report map[string]any, name string) string {
	t.Helper()
	digests, _ := report["digests"].(map[string]any)
	value, _ := digests[name].(string)
	if value == "" {
		t.Fatalf("the report carries no %s digest: %v", name, report["digests"])
	}
	return value
}

func planFromReport(t *testing.T, report map[string]any) map[string]any {
	t.Helper()
	run, _ := report["run"].(map[string]any)
	truth, _ := report["truth"].(map[string]any)
	start := run["start_time_ns"].(float64)
	secondsFrom := func(ns any) float64 { return math.Round((ns.(float64) - start) / 1e9) }
	calls := []any{}
	for _, raw := range truth["effector_calls"].([]any) {
		call := raw.(map[string]any)
		calls = append(calls, map[string]any{"entity_id": call["entity_id"], "effector": call["effector"], "args": call["args"], "at_s": secondsFrom(call["at_ns"])})
	}
	plan := map[string]any{"domain": canvasDomain, "seed": run["seed"], "duration_s": secondsFrom(run["requested_end_ns"]), "rate": 0,
		"faults": faultsFromReport(truth, secondsFrom), "perturbations": perturbationsFromReport(truth, secondsFrom), "calls": calls}
	for _, key := range []string{"name", "purpose"} {
		if text, ok := run[key]; ok {
			plan[key] = text
		}
	}
	return plan
}

func faultsFromReport(truth map[string]any, secondsFrom func(any) float64) []any {
	faults := []any{}
	for _, raw := range truth["faults"].([]any) {
		fault := raw.(map[string]any)
		planned := map[string]any{"entity_id": fault["entity_id"], "fault": fault["fault"], "at_s": secondsFrom(fault["onset_ns"])}
		if fault["cleared_by"] == "command" {
			planned["until_s"] = secondsFrom(fault["cleared_ns"])
		}
		faults = append(faults, planned)
	}
	return faults
}

func perturbationsFromReport(truth map[string]any, secondsFrom func(any) float64) []any {
	perturbations := []any{}
	for _, raw := range truth["perturbations"].([]any) {
		applied := raw.(map[string]any)
		if applied["action"] != "apply" {
			continue
		}
		planned := map[string]any{"kind": applied["perturbation"], "at_s": secondsFrom(applied["from_ns"])}
		if params, ok := applied["params"]; ok {
			planned["params"] = params
		}
		if until, _ := applied["until_ns"].(float64); until > 0 {
			planned["until_s"] = secondsFrom(until)
		}
		perturbations = append(perturbations, planned)
	}
	return perturbations
}
