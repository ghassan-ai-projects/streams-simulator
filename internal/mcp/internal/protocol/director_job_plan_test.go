package protocol

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func richJobArguments() map[string]any {
	args := jobArguments(0)
	args["name"], args["purpose"] = "Aerator loss and recovery", "Does oxygen recover once the aerator is back?"
	args["faults"] = []any{map[string]any{"entity_id": "site-a/pond-1", "fault": "aerator_failure", "at_s": 300, "until_s": 900}}
	args["perturbations"] = []any{map[string]any{"kind": "drop", "params": map[string]any{"rate": 0.4}, "at_s": 200, "until_s": 1200}}
	return args
}

func TestAJobWithANameARecoveryAndADeliveryConditionRunsAndReportsThemBack(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	started := mustCall(t, cs, "sim.job.start", richJobArguments())
	if started["name"] != "Aerator loss and recovery" || started["purpose"] == "" {
		t.Fatalf("job.start = %v, want the name and purpose echoed", started)
	}
	status := settledJob(t, cs, started["world_id"].(string))
	faults, _ := status["faults"].([]any)
	perturbations, _ := status["perturbations"].([]any)
	fault, _ := faults[0].(map[string]any)
	if status["state"] != "completed" || len(faults) != 1 || fault["until_s"] != float64(900) || fault["fault_id"] != "f-0" || len(perturbations) != 1 {
		t.Fatalf("job.status = %v, want the fault with its clear time and id, and the perturbation", status)
	}
}

func TestAJobPlanTheBinaryCannotRunIsRefusedWithoutAWorld(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		change func(map[string]any)
		want   string
	}{
		"unknown perturbation kind": {func(a map[string]any) {
			a["perturbations"] = []any{map[string]any{"kind": "no_such_kind", "at_s": 10}}
		}, "no_such_kind"},
		"perturbation parameter out of range": {func(a map[string]any) {
			a["perturbations"] = []any{map[string]any{"kind": "drop", "params": map[string]any{"rate": 7}, "at_s": 10}}
		}, "rate"},
		"clear before the fault begins": {func(a map[string]any) {
			a["faults"] = []any{map[string]any{"entity_id": "site-a/pond-1", "fault": "aerator_failure", "at_s": 300, "until_s": 100}}
		}, "until_s"},
		"name with a control character": {func(a map[string]any) { a["name"] = "bad\nname" }, "name"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
			args := richJobArguments()
			tc.change(args)
			if _, refusal := toolCall(t, cs, "sim.job.start", args); !strings.Contains(refusal, tc.want) {
				t.Fatalf("job.start = %q, want a refusal naming %q", refusal, tc.want)
			}
			if listed := mustCall(t, cs, "sim.job.list", map[string]any{}); len(listed["jobs"].([]any)) != 0 {
				t.Fatalf("job.list = %v, want no job after a refusal", listed)
			}
		})
	}
}

func TestThePlanSchemaClosesTheNewObjects(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	for name, change := range map[string]func(map[string]any){
		"perturbation without a kind": func(a map[string]any) { a["perturbations"] = []any{map[string]any{"at_s": 1}} },
		"perturbation with an extra key": func(a map[string]any) {
			a["perturbations"] = []any{map[string]any{"kind": "drop", "at_s": 1, "target": "x"}}
		},
		"perturbation time as text": func(a map[string]any) { a["perturbations"] = []any{map[string]any{"kind": "drop", "at_s": "soon"}} },
		"fraction of a second to clear": func(a map[string]any) {
			a["faults"] = []any{map[string]any{"entity_id": "x", "fault": "y", "at_s": 1, "until_s": 2.5}}
		},
		"name as a number": func(a map[string]any) { a["name"] = 5 },
	} {
		args := richJobArguments()
		change(args)
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "sim.job.start", Arguments: args})
		if err == nil && !res.IsError {
			t.Errorf("%s: job.start was accepted", name)
		}
	}
}

func TestThePerturbationCatalogToolDescribesEveryKindForTheDirectorOnly(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	listed, _ := mustCall(t, cs, "sim.catalog.perturbations", map[string]any{})["perturbations"].([]any)
	if len(listed) == 0 {
		t.Fatal("no perturbation kinds listed")
	}
	drop := listed[0].(map[string]any)
	for _, kind := range listed {
		if entry := kind.(map[string]any); entry["kind"] == "drop" {
			drop = entry
		}
	}
	params, _ := drop["params"].([]any)
	rate, _ := params[0].(map[string]any)
	if drop["category"] != "delivery" || rate["name"] != "rate" || rate["type"] != "number" || rate["maximum"] != float64(1) {
		t.Fatalf("drop = %v, want a delivery kind with its rate bound", drop)
	}
	if toolNames(t, NewOperatorServerResolver(d))["sim.catalog.perturbations"] {
		t.Fatal("the operator surface lists the perturbation catalog")
	}
}
