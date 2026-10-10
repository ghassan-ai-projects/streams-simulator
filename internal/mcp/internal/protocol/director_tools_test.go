package protocol

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// toolCall invokes one director tool over an in-memory session and returns its
// structured result, or the error text of a refused call.
func toolCall(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (map[string]any, string) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		for _, content := range res.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				return nil, text.Text
			}
		}
		return nil, "error without text"
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s result: %v", name, err)
	}
	return out, ""
}

func mustCall(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	out, refusal := toolCall(t, cs, name, args)
	if refusal != "" {
		t.Fatalf("%s refused: %s", name, refusal)
	}
	return out
}

func TestDirectorCatalogToolsDescribeTheInstalledDomainsAndAdapters(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	list := mustCall(t, cs, "sim.catalog.list", map[string]any{})
	if domains, _ := list["domains"].([]any); len(domains) != 1 {
		t.Fatalf("catalog list = %v", list)
	}
	described := mustCall(t, cs, "sim.catalog.describe", map[string]any{"domain": "aquaculture-pond"})
	if described["digest"] == "" || described["spec"] == nil {
		t.Fatalf("catalog describe = %v", described)
	}
	if _, refusal := toolCall(t, cs, "sim.catalog.describe", map[string]any{"domain": "no-such"}); !strings.HasPrefix(refusal, "domain_invalid: ") {
		t.Fatalf("unknown domain refusal = %q", refusal)
	}
	if coverage := mustCall(t, cs, "sim.catalog.coverage", map[string]any{}); coverage["by_axis"] == nil {
		t.Fatalf("coverage = %v", coverage)
	}
	adapters := mustCall(t, cs, "sim.adapter.list", map[string]any{})
	if list, _ := adapters["adapters"].([]any); len(list) != 1 {
		t.Fatalf("adapter list = %v", adapters)
	}
}

func TestDirectorWorldToolsCreateDescribeAdvanceAndDestroy(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	created := mustCall(t, cs, "sim.world.create", map[string]any{
		"domain": "aquaculture-pond", "seed": 5, "adapter": "native-jsonl", "sink": model.SinkInproc, "time_mode": model.TimeStepped,
	})
	worldID, _ := created["world_id"].(string)
	if worldID == "" || created["token"] == "" {
		t.Fatalf("created = %v", created)
	}
	described := mustCall(t, cs, "sim.world.describe", map[string]any{"world_id": worldID})
	if described["seed"] != float64(5) || described["emitted"] != float64(0) {
		t.Fatalf("describe = %v", described)
	}
	advanced := mustCall(t, cs, "sim.clock.advance", map[string]any{"world_id": worldID, "by_ns": 60 * 1e9})
	state := mustCall(t, cs, "sim.clock.state", map[string]any{"world_id": worldID})
	if state["clock"] != advanced["clock"] || state["pending_effects"] != float64(0) {
		t.Fatalf("clock state %v does not follow the advance %v", state, advanced)
	}
	if destroyed := mustCall(t, cs, "sim.world.destroy", map[string]any{"world_id": worldID}); destroyed["destroyed"] != true {
		t.Fatalf("destroy = %v", destroyed)
	}
	if _, refusal := toolCall(t, cs, "sim.world.describe", map[string]any{"world_id": worldID}); !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("describe after destroy = %q", refusal)
	}
}

func TestDirectorInjectionToolsInjectListAndClear(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	worldID := createWorld(t, d)
	w := d.World(worldID)
	entity := w.Run.World.EntityIDs()[0]
	fault := w.Run.Domain().Spec.Faults[0].ID

	injected := mustCall(t, cs, "sim.fault.inject", map[string]any{"world_id": worldID, "entity_id": entity, "fault": fault})
	faultID, _ := injected["fault_id"].(string)
	listed := mustCall(t, cs, "sim.fault.list", map[string]any{"world_id": worldID})
	if faults, _ := listed["faults"].([]any); len(faults) != 1 {
		t.Fatalf("fault list = %v", listed)
	}
	if cleared := mustCall(t, cs, "sim.fault.clear", map[string]any{"world_id": worldID, "fault_id": faultID}); cleared["cleared"] != true {
		t.Fatalf("fault clear = %v", cleared)
	}

	applied := mustCall(t, cs, "sim.perturb.apply", map[string]any{"world_id": worldID, "perturbation": "drop"})
	perturbID, _ := applied["perturb_id"].(string)
	if cleared := mustCall(t, cs, "sim.perturb.clear", map[string]any{"world_id": worldID, "perturb_id": perturbID}); cleared["cleared"] != true {
		t.Fatalf("perturb clear = %v", cleared)
	}
	_, refusal := toolCall(t, cs, "sim.env.inject", map[string]any{"world_id": worldID, "target": "mains", "fault": "outage"})
	if !strings.HasPrefix(refusal, "domain_invalid: ") || !strings.Contains(refusal, "env.inject not enabled") {
		t.Fatalf("env.inject without a configured target = %q", refusal)
	}
}

func TestDirectorTruthAndRunToolsEnforceTheSealOrder(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	worldID := createWorld(t, d)
	w := d.World(worldID)
	runID := w.Run.ID
	entity := w.Run.World.EntityIDs()[0]

	if _, refusal := toolCall(t, cs, "sim.run.begin", map[string]any{"world_id": worldID}); !strings.HasPrefix(refusal, "truth_sealed: ") {
		t.Fatalf("run.begin before sealing = %q", refusal)
	}
	sealed := mustCall(t, cs, "sim.truth.seal", map[string]any{"run_id": runID, "ground_truth": map[string]any{
		"scenario_id": "tools/0001", "domain": "aquaculture-pond", "seed": 42, "entity_id": entity,
		"label": "aerator_failure", "expected_episode": true, "injection_time_ns": 1, "first_observable_time_ns": 1,
		"unavoidable_time_ns": 2, "observability": map[string]any{"detector_form": "single_channel_snr", "channels": []any{"dissolved_oxygen"}, "effective_sigma": 0.1, "first_observable_snr": 3.0, "unavoidable_snr": 6.0}, "trivial_baseline_verdict": "non_trivial",
	}})
	if sealed["sealed"] != true {
		t.Fatalf("seal = %v", sealed)
	}
	status := mustCall(t, cs, "sim.truth.seal_status", map[string]any{"run_id": runID})
	if status["sealed"] != true {
		t.Fatalf("seal status = %v", status)
	}
	mustCall(t, cs, "sim.run.begin", map[string]any{"world_id": worldID, "label": "tools"})
	if _, refusal := toolCall(t, cs, "sim.truth.reveal", map[string]any{"run_id": runID}); refusal == "" {
		t.Fatal("truth must stay sealed while the run is open")
	}
	ended := mustCall(t, cs, "sim.run.end", map[string]any{"world_id": worldID})
	path, _ := ended["run_artifact_path"].(string)
	if verified := mustCall(t, cs, "sim.run.verify", map[string]any{"run_artifact_path": path}); verified["matches"] != true {
		t.Fatalf("verify = %v", verified)
	}
	revealed := mustCall(t, cs, "sim.truth.reveal", map[string]any{"run_id": runID})
	truth, _ := revealed["ground_truth"].(map[string]any)
	if truth["label"] != "aerator_failure" || truth["entity_id"] != entity {
		t.Fatalf("reveal after the run = %v", revealed)
	}
}

func TestDirectorEntityRetireAndAuditRefuseWhatTheyCannotDo(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	worldID := createWorld(t, d)
	entity := d.World(worldID).Run.World.EntityIDs()[0]

	retired := mustCall(t, cs, "sim.entity.retire", map[string]any{"world_id": worldID, "entity_id": entity, "reason": "decommissioned"})
	if retired["retired"] != true {
		t.Fatalf("retire = %v", retired)
	}
	_, refusal := toolCall(t, cs, "sim.entity.retire", map[string]any{"world_id": "w-nope", "entity_id": entity, "reason": "x"})
	if !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("retire in an unknown world = %q", refusal)
	}
	_, refusal = toolCall(t, cs, "sim.score", map[string]any{"run_id": "r-nope"})
	if refusal == "" {
		t.Fatal("scoring an unknown run must be refused")
	}
}

func TestDirectorResourcesServeTheCatalogAndDomainSpec(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	for _, uri := range []string{"sim://catalog", "sim://domains/aquaculture-pond/spec"} {
		res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
		if err != nil || len(res.Contents) != 1 || !strings.Contains(res.Contents[0].Text, "aquaculture-pond") {
			t.Fatalf("resource %s = %+v (%v)", uri, res, err)
		}
	}
	_, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "sim://domains/no-such/spec"})
	if err == nil || !strings.Contains(err.Error(), "no-such") {
		t.Fatalf("an unknown domain resource must be refused by name: %v", err)
	}
}

func TestAdvanceErrorsKeepTheirOwnCodes(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	worldID := createWorld(t, d)
	mustCall(t, cs, "sim.clock.advance", map[string]any{"world_id": worldID, "by_ns": 60 * 1e9})

	_, refusal := toolCall(t, cs, "sim.clock.advance", map[string]any{"world_id": worldID, "to_ns": model.DefaultStartTimeNS})
	if !strings.HasPrefix(refusal, "clock_backwards: ") {
		t.Fatalf("moving the clock back = %q", refusal)
	}
	mustCall(t, cs, "sim.world.destroy", map[string]any{"world_id": worldID})
	_, refusal = toolCall(t, cs, "sim.clock.advance", map[string]any{"world_id": worldID, "by_ns": 1})
	if !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("advancing a destroyed world = %q", refusal)
	}

	worldID = createWorld(t, d)
	w := d.World(worldID)
	if _, err := w.Run.End(""); err != nil {
		t.Fatal(err)
	}
	_, refusal = toolCall(t, cs, "sim.clock.advance", map[string]any{"world_id": worldID, "by_ns": 60 * 1e9})
	if !strings.HasPrefix(refusal, "domain_invalid: Advance: run is finished") {
		t.Fatalf("advancing a finished run must not claim the clock went backwards: %q", refusal)
	}
}

func TestWorldCreateRefusesSeedsThatJSONCannotCarryExactly(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	for name, seed := range map[string]any{"negative": -1, "beyond 2^53": float64(1 << 54), "fractional": 1.5} {
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "sim.world.create", Arguments: map[string]any{
			"domain": "aquaculture-pond", "seed": seed,
		}})
		if err == nil && !res.IsError {
			t.Errorf("%s seed %v was accepted", name, seed)
		}
	}
	created := mustCall(t, cs, "sim.world.create", map[string]any{"domain": "aquaculture-pond", "seed": float64(1<<53 - 1)})
	described := mustCall(t, cs, "sim.world.describe", map[string]any{"world_id": created["world_id"]})
	if described["seed"] != float64(1<<53-1) {
		t.Fatalf("the largest carried seed must round-trip exactly: %v", described["seed"])
	}
}

func TestWorldCreateReturnsTheRunIdTruthIsSealedAgainst(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	cs, _ := connect(t, NewDirectorServer(d))
	created := mustCall(t, cs, "sim.world.create", map[string]any{"domain": "aquaculture-pond"})
	runID, _ := created["run_id"].(string)
	if runID == "" || runID != d.World(created["world_id"].(string)).Run.ID {
		t.Fatalf("world.create must name the run: %v", created)
	}
}
