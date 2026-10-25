package acceptance

import (
	"testing"
)

func recoveryPlan() map[string]any {
	plan := jobPlan(240, 0)
	plan["name"], plan["purpose"] = "Fan failure and recovery", "Does the tower recover once its fan is restored while delivery drops messages?"
	plan["faults"].([]any)[0].(map[string]any)["until_s"] = 150
	plan["perturbations"] = []any{map[string]any{"kind": "drop", "params": map[string]any{"rate": 0.5}, "at_s": 30, "until_s": 200}}
	return plan
}

func TestANamedRunWithARecoveryAndADeliveryConditionReportsThemAndRepeatsExactly(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs two planned jobs through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	first := startJob(t, client, recoveryPlan())
	status := waitForJobState(t, client, first, "completed")
	requireStatusPlan(t, status)
	waitForListing(t, client, first, "the job to be listed complete", func(entry map[string]any) bool { return entry["complete"] == true })
	if listed := listedRun(t, client, first); listed["name"] != "Fan failure and recovery" {
		t.Fatalf("the listing says %v, want the run's name", listed)
	}
	report := client.action("run.report", map[string]any{"run": first})
	requireReportedPlan(t, report)
	again := startJob(t, client, planFromReport(t, report))
	waitForJobState(t, client, again, "completed")
	waitForListing(t, client, again, "the repeat to be listed complete", func(entry map[string]any) bool { return entry["complete"] == true })
	repeated := client.action("run.report", map[string]any{"run": again})
	for _, digest := range []string{"trace", "world", "ledger"} {
		if got, want := digestOf(t, repeated, digest), digestOf(t, report, digest); got != want {
			t.Errorf("the repeat's %s digest is %s, the original's is %s", digest, got, want)
		}
	}
	if verified := client.action("run.verify", map[string]any{"run": again}); verified["matches"] != true {
		t.Fatalf("verify of the repeat = %v", verified)
	}
}

func requireStatusPlan(t *testing.T, status map[string]any) {
	t.Helper()
	faults, _ := status["faults"].([]any)
	perturbations, _ := status["perturbations"].([]any)
	fault, _ := faults[0].(map[string]any)
	if status["name"] != "Fan failure and recovery" || len(faults) != 1 || fault["until_s"] != float64(150) || fault["fault_id"] != "f-0" || len(perturbations) != 1 {
		t.Fatalf("job.status = %v, want the name, the fault with its clear time and id, and the perturbation", status)
	}
}

func requireReportedPlan(t *testing.T, report map[string]any) {
	t.Helper()
	run, _ := report["run"].(map[string]any)
	truth, _ := report["truth"].(map[string]any)
	faults, _ := truth["faults"].([]any)
	perturbations, _ := truth["perturbations"].([]any)
	if run["name"] != "Fan failure and recovery" || run["purpose"] == nil || len(faults) != 1 || len(perturbations) != 1 {
		t.Fatalf("report run %v truth %v, want the name, purpose, one fault and one perturbation", run, truth)
	}
	fault, _ := faults[0].(map[string]any)
	start, onset, cleared := run["start_time_ns"].(float64), fault["onset_ns"].(float64), fault["cleared_ns"].(float64)
	if fault["cleared_by"] != "command" || cleared-start != 150e9 || onset-start != 60e9 {
		t.Fatalf("fault %v, want it cleared by command at 150 s after a 60 s onset", fault)
	}
	applied, _ := perturbations[0].(map[string]any)
	if applied["perturbation"] != "drop" || applied["from_ns"].(float64)-start != 30e9 || applied["until_ns"].(float64)-start != 200e9 {
		t.Fatalf("perturbation %v, want drop from 30 s to 200 s", applied)
	}
	if counts, _ := report["counts"].(map[string]any); counts["dropped"] == nil || counts["dropped"].(float64) == 0 {
		t.Fatalf("counts %v, want the dropped deliveries the perturbation caused", counts)
	}
}

func TestTheBinaryListsTheDeliveryPerturbationsAndTheCategoryOfEachFault(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and asks its catalogs through the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	catalog, _ := client.action("catalog.perturbations", nil)["perturbations"].([]any)
	if len(catalog) == 0 {
		t.Fatal("no perturbation kinds listed")
	}
	described := client.action("catalog.describe", map[string]any{"domain": canvasDomain})
	categories, _ := described["fault_categories"].(map[string]any)
	tower, _ := categories["cooling_tower"].(map[string]any)
	if tower["cooling_tower__fan_failure"] != "world" {
		t.Fatalf("fault categories = %v, want the cooling tower's fan failure as a world fault", categories)
	}
}
