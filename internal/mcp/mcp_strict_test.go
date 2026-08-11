package mcp

// Slice A: the nested contract schemas (verdict, ground_truth) are typed and
// enforced over the wire; fault/perturb params reject unknown keys; start_time
// is presence-aware so an explicit epoch-0 start is honored.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// schemaOf returns the advertised InputSchema of one tool.
func schemaOf(t *testing.T, server *mcp.Server, name string) map[string]any {
	t.Helper()
	cs, _ := connect(t, server)
	res, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if tool.Name == name {
			raw, err := json.Marshal(tool.InputSchema)
			if err != nil {
				t.Fatalf("marshal schema for %s: %v", name, err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("unmarshal schema for %s: %v", name, err)
			}
			return doc
		}
	}
	t.Fatalf("tool %s not advertised", name)
	return nil
}

func TestNestedContractSchemasAreAdvertised(t *testing.T) {
	// ground_truth rides the director surface; verdict rides the operator
	// surface. Each server advertises its own typed nested contract.
	check := func(t *testing.T, server *mcp.Server, tool, prop string) {
		t.Helper()
		doc := schemaOf(t, server, tool)
		props, ok := doc["properties"].(map[string]any)
		if !ok {
			t.Fatalf("%s: no properties", tool)
		}
		nested, ok := props[prop].(map[string]any)
		if !ok {
			t.Fatalf("%s: no %s property", tool, prop)
		}
		if nested["type"] != "object" {
			t.Fatalf("%s.%s: expected type object, got %v", tool, prop, nested["type"])
		}
		if _, ok := nested["properties"].(map[string]any); !ok {
			t.Fatalf("%s.%s: expected typed properties, got %v", tool, prop, nested)
		}
		if _, ok := nested["required"].([]any); !ok {
			t.Fatalf("%s.%s: expected a required list", tool, prop)
		}
		if nested["additionalProperties"] != false {
			t.Fatalf("%s.%s: expected additionalProperties false, got %v", tool, prop, nested["additionalProperties"])
		}
	}
	d := newTestDirector(t)
	check(t, NewDirectorServer(d), "sim.truth.seal", "ground_truth")
	d2 := newTestDirector(t)
	worldID := createWorld(t, d2)
	check(t, NewOperatorServer(d2.Worlds[worldID].Operator), "sim.consumer.report", "verdict")
}

func validVerdict(runID string) map[string]any {
	return map[string]any{
		"schema_version": "0.1",
		"run_id":         runID,
		"consumer":       map[string]any{"name": "strict-test", "version": "0.1"},
	}
}

func TestVerdictContractEnforcedOverWire(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	runID := d.Worlds[worldID].Run.ID
	w := d.Worlds[worldID]
	cs, _ := connect(t, NewOperatorServer(w.Operator))

	report := func(args map[string]any) (*mcp.CallToolResult, error) {
		return cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "sim.consumer.report", Arguments: args,
		})
	}

	// A structurally valid verdict is accepted.
	res, err := report(map[string]any{
		"token": w.Token, "run_id": runID,
		"quiesced_through_ns": int64(0), "verdict": validVerdict(runID),
	})
	if err != nil {
		t.Fatalf("valid verdict rejected: %v", err)
	}
	if res.IsError {
		t.Fatalf("valid verdict rejected: %v", res)
	}

	// An unknown top-level verdict key violates the contract.
	bad := validVerdict(runID)
	bad["bogus_field"] = 1
	res, err = report(map[string]any{"token": w.Token, "run_id": runID, "verdict": bad})
	if err == nil && !res.IsError {
		t.Fatal("verdict with unknown top-level key must be rejected")
	}

	// A missing required field is rejected.
	missing := validVerdict(runID)
	delete(missing, "consumer")
	res, err = report(map[string]any{"token": w.Token, "run_id": runID, "verdict": missing})
	if err == nil && !res.IsError {
		t.Fatal("verdict missing consumer must be rejected")
	}
}

func TestGroundTruthContractEnforcedOverWire(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	runID := d.Worlds[worldID].Run.ID
	cs, _ := connect(t, NewDirectorServer(d))

	seal := func(gt map[string]any) (*mcp.CallToolResult, error) {
		return cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "sim.truth.seal",
			Arguments: map[string]any{"run_id": runID, "ground_truth": gt},
		})
	}

	// Incomplete record: missing the required label.
	rec := map[string]any{
		"scenario_id": "strict/0001", "domain": "aquaculture-pond", "seed": int64(1),
		"entity_id":         "site-a/pond-1",
		"expected_episode":  true,
		"injection_time_ns": int64(0), "first_observable_time_ns": int64(0),
		"unavoidable_time_ns": int64(0),
		"observability": map[string]any{
			"detector_form": "single_channel_snr", "channels": []any{"do_saturation"},
			"effective_sigma": 0.1, "first_observable_snr": 3.0, "unavoidable_snr": 6.0,
		},
		"trivial_baseline_verdict": "non_trivial",
	}
	res, err := seal(rec)
	if err == nil && !res.IsError {
		t.Fatal("incomplete ground truth must be rejected")
	}

	// An unknown top-level key violates the contract.
	rec["bogus"] = "x"
	res, err = seal(rec)
	if err == nil && !res.IsError {
		t.Fatal("ground truth with unknown key must be rejected")
	}
}

func TestFaultInjectRejectsUnknownParamKeys(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	cs, _ := connect(t, NewDirectorServer(d))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "sim.fault.inject",
		Arguments: map[string]any{
			"world_id": worldID, "entity_id": "site-a/pond-1", "fault": "aerator_failure",
			"params": map[string]any{"severity": 2.0, "bogus": 1},
		},
	})
	if err == nil && !res.IsError {
		t.Fatal("fault.inject with unknown param key must be rejected")
	}
}

func TestPerturbApplyRejectsUnknownParamKeys(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	cs, _ := connect(t, NewDirectorServer(d))
	for _, params := range []map[string]any{
		{"raet": 0.5},             // misspelled rate
		{"rate": "high"},          // wrong type
		{"rate": 0.5, "extra": 1}, // unknown key alongside a declared one
	} {
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "sim.perturb.apply",
			Arguments: map[string]any{
				"world_id": worldID, "perturbation": "drop", "params": params,
			},
		})
		if err == nil && !res.IsError {
			t.Fatalf("perturb.apply with params %v must be rejected", params)
		}
	}
}

func TestWorldCreateHonorsExplicitZeroStartTime(t *testing.T) {
	d := newTestDirector(t)
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(1), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
		"start_time": float64(0),
	})
	if err != nil {
		t.Fatalf("explicit epoch-0 start rejected: %v", err)
	}
	if got := res["clock"]; got != "1970-01-01T00:00:00Z" {
		t.Fatalf("expected epoch-0 clock, got %v", got)
	}

	// Absent start_time keeps the documented default.
	res, err = d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(2), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := res["clock"]; got != "2026-01-01T00:00:00Z" {
		t.Fatalf("expected default start, got %v", got)
	}
}
