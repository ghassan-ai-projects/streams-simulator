package protocol

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func rawToolResult(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func blindJobArguments(explore bool) map[string]any {
	args := jobArguments(1)
	args["explore"] = explore
	args["perturbations"] = []any{map[string]any{"kind": "drop", "params": map[string]any{"rate": 0.3}, "at_s": 100}}
	return args
}

type toolRequest struct {
	name string
	args map[string]any
}

func controlRequests(worldID any) []toolRequest {
	world := map[string]any{"world_id": worldID}
	pond := "site-a/pond-1"
	return []toolRequest{{"sim.job.status", world}, {"sim.job.list", map[string]any{}}, {"sim.job.pause", world}, {"sim.job.step", world},
		{"sim.job.rate", map[string]any{"world_id": worldID, "rate": 1}},
		{"sim.job.invoke_effector", map[string]any{"world_id": worldID, "entity_id": pond, "effector": "start_aerator", "args": map[string]any{"pond_id": pond}}},
		{"sim.job.inject_fault", map[string]any{"world_id": worldID, "entity_id": pond, "fault": "do_probe_fouling"}},
		{"sim.job.resume", world}, {"sim.job.stop", world}}
}

func TestJobToolsOfAnEvaluateJobCarryNoPlanOnTheWireWhileItIsActive(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", blindJobArguments(false))
	for _, request := range controlRequests(started["world_id"]) {
		raw := rawToolResult(t, cs, request.name, request.args)
		for _, secret := range []string{"aerator_failure", "drop", "do_probe_fouling", "fault_id"} {
			if strings.Contains(raw, secret) {
				t.Fatalf("%s result %s carries %q while the evaluate job is active", request.name, raw, secret)
			}
		}
	}
	ended := settledJob(t, cs, started["world_id"].(string))
	if faults, _ := ended["faults"].([]any); len(faults) != 1 {
		t.Fatalf("ended status = %v, want the plan shown once the job ended", ended)
	}
}

func TestJobInjectFaultRefusesAnEvaluateJobWithEvaluateBlind(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", blindJobArguments(false))
	request := controlRequests(started["world_id"])[6]
	if _, refusal := toolCall(t, cs, request.name, request.args); !strings.HasPrefix(refusal, "evaluate_blind: ") {
		t.Fatalf("refusal = %q, want evaluate_blind", refusal)
	}
	mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": started["world_id"]})
	settledJob(t, cs, started["world_id"].(string))
}

func TestJobToolsOfAnExploreJobShowThePlanWhileItIsActive(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", blindJobArguments(true))
	if faults, _ := started["faults"].([]any); len(faults) != 1 {
		t.Fatalf("start = %v, want the plan of an explore job", started)
	}
	status := mustCall(t, cs, "sim.job.status", map[string]any{"world_id": started["world_id"]})
	if perturbations, _ := status["perturbations"].([]any); len(perturbations) != 1 {
		t.Fatalf("status = %v, want the delivery conditions of an explore job", status)
	}
	mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": started["world_id"]})
	settledJob(t, cs, started["world_id"].(string))
}

func TestJobToolsOfEitherModeListThePlannedCallsAndAnExploreJobKeepsAHandFaultOnTheWire(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{false, true} {
		cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
		started := mustCall(t, cs, "sim.job.start", blindJobArguments(explore))
		id := started["world_id"]
		status := mustCall(t, cs, "sim.job.status", map[string]any{"world_id": id})
		planned, _ := status["planned_calls"].([]any)
		first, _ := planned[0].(map[string]any)
		if len(planned) != 1 || first["effector"] != "start_aerator" || first["at_s"] != float64(900) || first["entity_id"] != "site-a/pond-1" {
			t.Fatalf("explore=%v status = %v, want the one planned call with its time", explore, status)
		}
		if explore {
			assertHandFaultListed(t, cs, id)
		}
		mustCall(t, cs, "sim.job.stop", map[string]any{"world_id": id})
		settledJob(t, cs, id.(string))
	}
}

func assertHandFaultListed(t *testing.T, cs *mcp.ClientSession, id any) {
	t.Helper()
	mustCall(t, cs, "sim.job.pause", map[string]any{"world_id": id})
	injected := mustCall(t, cs, "sim.job.inject_fault", map[string]any{"world_id": id, "entity_id": "site-a/pond-1", "fault": "do_probe_fouling"})
	faults, _ := mustCall(t, cs, "sim.job.status", map[string]any{"world_id": id})["faults"].([]any)
	hand, _ := faults[len(faults)-1].(map[string]any)
	if len(faults) != 2 || hand["immediate"] != true || hand["fault"] != "do_probe_fouling" || hand["fault_id"] != injected["fault_id"] || hand["at_s"] == nil {
		t.Fatalf("faults = %v, want the hand fault listed with its id, second and immediate flag", faults)
	}
}
