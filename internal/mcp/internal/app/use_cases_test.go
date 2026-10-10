package app

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestWorldCommandsDriveAWorldWithoutAnyProtocol(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.World(worldID)
	entity := w.Run.World.EntityIDs()[0]
	fault := w.Run.Domain().Spec.Faults[0].ID

	if described, err := d.DescribeWorld(worldID); err != nil || described["world_id"] != worldID {
		t.Fatalf("describe = %v, %v", described, err)
	}
	if state, err := d.ClockState(worldID); err != nil || state["pending_effects"] == nil {
		t.Fatalf("clock state = %v, %v", state, err)
	}
	injected, err := d.InjectFault(worldID, entity, fault, w.Run.World.Clock()+60e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	faultID, _ := injected["fault_id"].(string)
	if listed, err := d.ListFaults(worldID); err != nil || len(w.Run.World.ListFaults()) != 1 || listed["faults"] == nil {
		t.Fatalf("faults = %v, %v", listed, err)
	}
	if cleared, err := d.ClearFault(worldID, faultID); err != nil || cleared["cleared"] != true {
		t.Fatalf("clear fault = %v, %v", cleared, err)
	}
	applied, err := d.ApplyPerturb(worldID, "drop", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cleared, err := d.ClearPerturb(worldID, applied["perturb_id"].(string)); err != nil || cleared["cleared"] != true {
		t.Fatalf("clear perturb = %v, %v", cleared, err)
	}
	if _, err := d.EnvInject(worldID, "mains", "outage", nil); err == nil || !strings.Contains(err.Error(), "env.inject not enabled") {
		t.Fatalf("env inject without a target: %v", err)
	}
}

func TestWorldCommandsNameAnUnknownWorld(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	calls := map[string]func() error{
		"describe": func() error { _, err := d.DescribeWorld("w-9"); return err },
		"clock":    func() error { _, err := d.ClockState("w-9"); return err },
		"faults":   func() error { _, err := d.ListFaults("w-9"); return err },
		"clear":    func() error { _, err := d.ClearFault("w-9", "f-0"); return err },
		"perturb":  func() error { _, err := d.ApplyPerturb("w-9", "drop", nil, 0, 0); return err },
	}
	for name, call := range calls {
		var tool *ToolError
		if err := call(); !errors.As(err, &tool) || tool.Code != CodeWorldNotFound {
			t.Errorf("%s on an unknown world: %v", name, err)
		}
	}
}

func TestSealStatusAndVerifyRunFollowTheRunLifecycle(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.World(worldID)
	if _, err := d.SealStatus(w.Run.ID); err == nil || !strings.Contains(err.Error(), "unknown run") {
		t.Fatalf("an unsealed run has no seal status: %v", err)
	}
	if err := d.SealTruth(w.Run.ID, &model.GroundTruthRecord{ScenarioID: "u/1", Domain: "aquaculture-pond", Label: "l", EntityID: "e"}); err != nil {
		t.Fatal(err)
	}
	again := d.SealTruth(w.Run.ID, &model.GroundTruthRecord{ScenarioID: "u/2", Domain: "aquaculture-pond", Label: "l", EntityID: "e"})
	if again == nil || !strings.Contains(again.Error(), "already sealed") {
		t.Fatalf("sealing twice must be refused: %v", again)
	}
	if status, err := d.SealStatus(w.Run.ID); err != nil || status["sealed"] != true {
		t.Fatalf("status = %v, %v", status, err)
	}
	if _, err := d.BeginRun(worldID, "u"); err != nil {
		t.Fatal(err)
	}
	ended, err := d.EndRun(worldID)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := d.VerifyRun(ended["run_artifact_path"].(string))
	if err != nil || verified["matches"] != true {
		t.Fatalf("verify = %v, %v", verified, err)
	}
	if _, err := d.VerifyRun(t.TempDir() + "/absent.json"); err == nil || !strings.Contains(err.Error(), "absent.json") {
		t.Fatalf("verifying a missing artifact must name it: %v", err)
	}
}

func TestArgumentReadersAcceptTheNumberFormsAJSONClientSends(t *testing.T) {
	t.Parallel()
	args := map[string]any{
		"float": float64(7), "int": int64(8), "number": json.Number("9"), "text": "10", "word": "x", "fraction": 1.5,
	}
	for key, want := range map[string]int64{"float": 7, "int": 8, "number": 9, "text": 10} {
		if got, ok := IntArg(args, key); !ok || got != want {
			t.Errorf("IntArg(%s) = %d, %v; want %d", key, got, ok, want)
		}
	}
	if _, ok := IntArg(args, "word"); ok {
		t.Error("a non-numeric string is not an integer")
	}
	if got := Num(args, "absent", 42); got != 42 {
		t.Errorf("Num default = %d", got)
	}
	if got := Str(args, "text"); got != "10" {
		t.Errorf("Str = %q", got)
	}
}
