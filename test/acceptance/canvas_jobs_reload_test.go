package acceptance

import (
	"testing"
)

func TestAReloadedBrowserFindsTheCallsAndTheHandFaultsOfAnExploreJobAndItsRunRowIsNamed(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and reloads the browser of a paced explore job")
	}
	client, bridge, _ := pairedJobBridge(t)
	plan := jobPlan(1800, 10)
	plan["explore"], plan["name"] = true, "Fan loss"
	world := startJob(t, client, plan)
	args := map[string]any{"world_id": world}
	client.action("job.pause", args)
	hand := client.action("job.inject_fault", map[string]any{"world_id": world, "entity_id": "CT-1", "fault": canvasFault})
	reloaded := reloadedBrowser(t, client, bridge)
	status := jobState(t, reloaded, world)
	checkPlannedFanCall(t, status, 225)
	checkHandFault(t, status, hand["fault_id"])
	checkRunRowOfARunningJob(t, listedRun(t, reloaded, world))
	client.action("job.resume", args)
	client.action("job.stop", args)
	waitForJobState(t, reloaded, world, "stopped")
	row := listedRun(t, reloaded, world)
	if row["name"] != "Fan loss" || row["plant"] != canvasDomain || row["explore"] != true || row["running"] == true {
		t.Fatalf("the ended run row says %v, want the recorded name, plant and mode", row)
	}
}

func checkPlannedFanCall(t *testing.T, status map[string]any, atS float64) {
	t.Helper()
	planned, _ := status["planned_calls"].([]any)
	if len(planned) != 1 {
		t.Fatalf("the reloaded job says %v, want its one planned call", status["planned_calls"])
	}
	call, _ := planned[0].(map[string]any)
	args, _ := call["args"].(map[string]any)
	if call["entity_id"] != "CT-1" || call["effector"] != canvasEffector || call["at_s"] != atS || args["speed"] != 0.8 {
		t.Fatalf("planned call = %v, want the plan's own call", call)
	}
}

func checkHandFault(t *testing.T, status map[string]any, faultID any) {
	t.Helper()
	faults, _ := status["faults"].([]any)
	if len(faults) != 2 {
		t.Fatalf("the reloaded explore job says %v, want the planned fault and the hand fault", status["faults"])
	}
	hand, _ := faults[1].(map[string]any)
	if hand["immediate"] != true || hand["fault"] != canvasFault || hand["fault_id"] != faultID || hand["at_s"] == nil {
		t.Fatalf("hand fault = %v, want it listed as immediate with its id and second", hand)
	}
}

func checkRunRowOfARunningJob(t *testing.T, row map[string]any) {
	t.Helper()
	if row["running"] != true || row["name"] != "Fan loss" || row["plant"] != canvasDomain || row["explore"] != true {
		t.Fatalf("the running job's row says %v, want its name, plant and mode", row)
	}
}
