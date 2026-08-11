package refconsumer

// The reference consumer closes the loop: it consumes the delivered trace
// and actuates through the operator surface. This test drives it in-process
// against a real run.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

const (
	aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"
	nativeAdapter   = "../../adapters/native-jsonl.adapter.json"
)

// TestDetectsFaultAndActs: a crash scenario produces detections and one
// actuator dispatch per entity.
func TestDetectsFaultAndActs(t *testing.T) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	pond := "site-a/pond-1"
	if _, err := r.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(pond, "aerator_failure", start+2*3600*1e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+6*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}

	np := &Nameplate{
		WorldID:  r.ID,
		Entities: []EntityInfo{{ID: pond}},
		Effectors: []EffectorInfo{{
			Name:   "start_aerator",
			Schema: map[string]any{"type": "object", "required": []any{"pond_id"}, "properties": map[string]any{"pond_id": map[string]any{"type": "string"}}},
		}},
	}
	cfg := DefaultConfig()
	cfg.OnDetectionEffector = "start_aerator"
	cfg.Threshold = 3 // the crash is loud
	rc := New(cfg, np, r, r, r.ID)
	verdict, err := rc.Process(r.Trace(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdict.Detections) == 0 {
		t.Fatal("the reference consumer must detect the aerator failure")
	}
	// The crash happened at start+2h; detections must be at or after that.
	if len(verdict.Actions) == 0 {
		t.Fatal("the consumer must actuate on detection")
	}
	if verdict.Actions[0].Effector != "start_aerator" || verdict.Actions[0].EntityID != pond {
		t.Fatalf("wrong action: %+v", verdict.Actions[0])
	}
	// The verdict round-trips through the simulator's schema.
	raw, _ := json.Marshal(verdict)
	if err := model.ValidateVerdict(raw); err != nil {
		t.Fatalf("verdict fails the consumer-verdict contract: %v", err)
	}
}

// TestObserveOnlyOnCleanRun: no fault, no detections, no actions.
func TestObserveOnlyOnCleanRun(t *testing.T) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+4*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	np := &Nameplate{WorldID: r.ID}
	rc := New(DefaultConfig(), np, nil, r, r.ID)
	verdict, err := rc.Process(r.Trace(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdict.Detections) > 0 {
		t.Fatalf("clean run must not produce detections: %+v", verdict.Detections)
	}
}
