package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func TestCapabilityTokenEnforced(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.Worlds[worldID]
	// Without the token every operator call is refused.
	if _, err := w.Operator.ReadNameplate(""); err == nil {
		t.Fatal("nameplate.read without token must fail")
	}
	np, err := w.Operator.ReadNameplate(w.Token)
	if err != nil {
		t.Fatalf("nameplate.read with token failed: %v", err)
	}
	if np.Domain != "aquaculture-pond" || len(np.Channels) == 0 || len(np.Effectors) == 0 {
		t.Fatalf("nameplate incomplete: %+v", np)
	}
	// Unknown effector refused with a stable code.
	_, err = w.Operator.Invoke(w.Token, "no_such_effector", "site-a/pond-1", "cmd-x", nil, 0)
	var te *ToolError
	if !errors.As(err, &te) || te.Code != CodeUnknownEffector {
		t.Fatalf("expected unknown_effector, got %v", err)
	}
	// Missing command_id refused.
	_, err = w.Operator.Invoke(w.Token, "start_aerator", "site-a/pond-1", "", nil, 0)
	if err == nil {
		t.Fatal("missing command_id must fail")
	}
}

func TestDirectorCreatesDistinctWorldsAndOpaqueTokens(t *testing.T) {
	d := newTestDirector(t)
	one := createWorld(t, d)
	two := createWorld(t, d)
	if one == two {
		t.Fatalf("world ids collided: %q", one)
	}
	w1, w2 := d.Worlds[one], d.Worlds[two]
	if w1 == nil || w2 == nil || w1.Run.ID == w2.Run.ID {
		t.Fatalf("run ids collided: %q and %q", w1.Run.ID, w2.Run.ID)
	}
	if w1.Token == w2.Token || len(w1.Token) < 40 || len(w2.Token) < 40 {
		t.Fatalf("capability tokens are not opaque and unique: %q %q", w1.Token, w2.Token)
	}
	if w1.Token == fmt.Sprintf("t-%d-%d", 1, 42) {
		t.Fatal("capability token still exposes the predictable seed form")
	}
}

func TestBeginRunRequiresSealedTruth(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	if _, err := d.BeginRun(worldID, "missing-truth"); err == nil {
		t.Fatal("run.begin must refuse an unsealed oracle")
	}
}

func TestDirectorPassesSinkTarget(t *testing.T) {
	d := newTestDirector(t)
	target := t.TempDir() + "/trace.jsonl"
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(9), "adapter": "native-jsonl",
		"sink": model.SinkFile, "sink_target": target,
	})
	if err != nil {
		t.Fatal(err)
	}
	worldID := res["world_id"].(string)
	if _, err := d.Advance(context.Background(), worldID, model.DefaultStartTimeNS+60*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("director did not pass sink target through: %v", err)
	}
}

func TestClosedLoopThroughMCPSurface(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.Worlds[worldID]
	start := model.DefaultStartTimeNS + 4*3600*1e9
	pond := "site-a/pond-1"

	// Scenario context: aerator on at night, then a fault.
	if _, err := w.Run.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(context.Background(), worldID, start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.InjectFault(worldID, pond, "aerator_failure", 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(context.Background(), worldID, start+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	// The operator invokes the effector through the narrow surface.
	res, err := w.Operator.Invoke(w.Token, "start_aerator", pond, "cmd-mcp-1",
		map[string]any{"pond_id": pond, "level": 1.0}, w.Run.World.Clock())
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if !res.Simulated || !res.Accepted {
		t.Fatalf("invoke result wrong: %+v", res)
	}
	if _, err := d.Advance(context.Background(), worldID, start+7*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	// Submit a verdict through the report tool.
	verdict := &model.Verdict{
		SchemaVersion: "0.1", RunID: w.Run.ID,
		Consumer: model.ConsumerInfo{Name: "test", Version: "0.1"},
		Actions: []model.Action{{
			CommandID: "cmd-mcp-1", Effector: "start_aerator", EntityID: pond,
			IssuedAt: model.FormatTime(w.Run.World.Clock()), OutcomeBelieved: model.BelievedSucceeded,
		}},
	}
	if err := w.Operator.Report(w.Token, w.Run.ID, w.Run.World.Clock(), verdict); err != nil {
		t.Fatal(err)
	}
	// Seal the director-only label before opening the run. The operator never
	// receives this record.
	rec := &model.GroundTruthRecord{
		ScenarioID: "mcp/0001", Domain: "aquaculture-pond", Label: "aerator_failure",
		EntityID: pond, ExpectedEffector: "start_aerator",
		InjectionTimeNS: start + 2*3600*1e9, FirstObservableTimeNS: start + 2*3600*1e9,
	}
	if err := d.SealTruth(w.Run.ID, rec); err != nil {
		t.Fatal(err)
	}
	// Run lifecycle + score through the director.
	if _, err := d.BeginRun(worldID, "mcp-loop"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.EndRun(worldID); err != nil {
		t.Fatal(err)
	}
	out, err := d.Score(w.Run.ID)
	if err != nil {
		t.Fatal(err)
	}
	sc, ok := out["scorecard"].(*score.Scorecard)
	if !ok {
		t.Fatalf("scorecard missing: %T", out["scorecard"])
	}
	if !sc.Loop.Resolved {
		t.Fatalf("loop did not resolve: %+v", sc.Loop)
	}
}
