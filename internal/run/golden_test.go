package run

// S1 gate 2: the hand-computed golden vector. A minimal world with no
// noise, no jitter and a constant F0 state produces events a human can
// verify: every field below (times, values, seq, world id) was computed by
// hand, not by the simulator. If this test passes, the emission pipeline is
// right in the small; the analytic cross-check covers the large.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// goldenSpec: state v = baseline 5.0 (no trend, no seasonality, no drift);
// channel sensor reads 2*v + 0.5 = 10.5 exactly, quantized at 0.1, emitted
// every 60 s with a constant 1.5 s link delay.
func goldenSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	doc := `{
		"id": "golden", "version": "0.1.0", "title": "golden vector domain",
		"description": "A constant-state domain used to hand-verify the emission pipeline.",
		"stresses": "` + fmt.Sprintf("%040d", 0) + `",
		"axes": {"rate": "low", "cardinality": "singleton", "value_shape": ["scalar"], "cadence": ["periodic"], "lateness": "bounded", "absence": "signal", "time_reference": "wall", "correlation": ["independent"], "seasonality": ["none"], "actuation": "observe_only", "consequence": ["cost"], "fidelity": ["F0"]},
		"entities": {"id_template": "golden-{n}", "count": {"default": 1}},
		"state": [{"name": "v", "unit": "u", "initial": 5}],
		"dynamics": [{"target": "v", "tier": "F0", "f0": {"baseline": 5}}],
		"channels": [
			{"name": "sensor", "description": "constant sensor", "value_type": "number", "unit": "u", "resolution": 0.1, "range": {"min": 0, "max": 100}, "observes": "v", "observation_gain": 2, "observation_offset": 0.5, "fidelity": "F0", "absence": "signal", "cadence": {"mode": "periodic", "period_s": 60}, "noise": {"model": "gaussian", "sigma": 0}, "link_delay": {"model": "constant", "mean_s": 1.5}}
		],
		"faults": [{"id": "f1", "description": "unused", "onset": {"shape": "step"}, "affects": [{"state": "v", "delta": 1}], "observability": {"detector": {"form": "single_channel_snr", "channel": "sensor"}}}],
		"profiles": [
			{"name": "nominal", "description": "x"},
			{"name": "correlated_cascade", "description": "x", "not_applicable": "` + fmt.Sprintf("%020d", 0) + `"},
			{"name": "sensor_pathology", "description": "x"}
		],
		"ground_truth": {"negative_class_fraction": 0.4}
	}`
	c, err := domain.Parse([]byte(doc), "golden")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestHandComputedGoldenVector(t *testing.T) {
	spec := goldenSpec(t)
	start := model.DefaultStartTimeNS // 2026-01-01T00:00:00Z
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 1, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, RunID: "r-golden",
		WorldID: "w-hand",
	})
	if err != nil {
		t.Fatal(err)
	}
	// The world id must be the fixed one the hand-written trace uses.
	if r.World.ID != "w-hand" {
		t.Fatalf("world id is %q, want w-hand", r.World.ID)
	}
	if _, err := r.Advance(context.Background(), start+12*60*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(r.trace), "\n"), "\n")
	if len(lines) != 12 {
		t.Fatalf("expected 12 events, got %d", len(lines))
	}
	for k := 0; k < 12; k++ {
		var ev model.SimEvent
		if err := json.Unmarshal([]byte(lines[k]), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.Seq != int64(k) {
			t.Fatalf("event %d: seq %d", k, ev.Seq)
		}
		if ev.WorldID != "w-hand" || ev.EntityID != "golden-1" || ev.Channel != "sensor" {
			t.Fatalf("event %d identity wrong: %+v", k, ev)
		}
		if ev.Value != 10.5 {
			t.Fatalf("event %d: value %v, want 10.5 (2*5 + 0.5)", k, ev.Value)
		}
		if ev.Unit != "u" {
			t.Fatalf("event %d: unit %q", k, ev.Unit)
		}
		et, _ := model.ParseTime(ev.EventTime)
		wantET := start + int64(k+1)*60*1e9
		if et != wantET {
			t.Fatalf("event %d: event_time %d, want %d (start + %ds)", k, et, wantET, (k+1)*60)
		}
		ot, _ := model.ParseTime(ev.ObservedTime)
		if ot != wantET+1500000000 {
			t.Fatalf("event %d: observed_time %d, want %d (+1.5s)", k, ot, wantET+1500000000)
		}
	}
	// The first record, spelled out: exactly what a human computes.
	want0 := `{"seq":0,"world_id":"w-hand","entity_type":"golden","entity_id":"golden-1","channel":"sensor","event_time":"2026-01-01T00:01:00Z","observed_time":"2026-01-01T00:01:01.5Z","value":10.5,"unit":"u"}`
	if lines[0] != want0 {
		t.Fatalf("first record differs:\n got %s\nwant %s", lines[0], want0)
	}
}
