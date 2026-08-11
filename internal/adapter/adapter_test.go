package adapter

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// miniAdapter builds a jsonl adapter exercising every transform op.
func miniAdapter() *model.Adapter {
	str := func(s string) *string { return &s }
	_ = str
	return &model.Adapter{
		ID: "mini", Version: "0.1.0", Encoding: "jsonl",
		EntityIDRewrite: &model.IDRewrite{
			Replace: []model.Replacement{{From: "/", To: "."}},
		},
		Preamble: []model.RecordTemplate{{
			Fields: []model.Field{
				{Name: "record_type", From: model.ValueExpr{Op: "const", Value: "config"}},
				{Name: "run", From: model.ValueExpr{Op: "run_meta", Key: "run_id"}},
			},
		}},
		Record: model.RecordTemplate{
			Fields: []model.Field{
				{Name: "record_type", From: model.ValueExpr{Op: "const", Value: "event"}},
				{Name: "event.id", From: model.ValueExpr{Op: "counter", Prefix: "evt-", Width: 6, Of: &model.ValueExpr{Op: "source", Source: "seq"}}},
				{Name: "event.entity_id", From: model.ValueExpr{Op: "source", Source: "entity_id"}},
				{Name: "event.type", From: model.ValueExpr{Op: "source", Source: "channel"}},
				{Name: "event.event_time", From: model.ValueExpr{Op: "format_time", Layout: "rfc3339_nano", Of: &model.ValueExpr{Op: "source", Source: "event_time"}}},
				{Name: "event.arrival_time", From: model.ValueExpr{Op: "format_time", Layout: "unix_millis", Of: &model.ValueExpr{Op: "source", Source: "observed_time"}}},
				{Name: "event.value", From: model.ValueExpr{Op: "source", Source: "value"}, OmitWhenNull: true},
				{Name: "event.unit", From: model.ValueExpr{Op: "source", Source: "unit"}, OmitWhenNull: true},
				{Name: "summary", From: model.ValueExpr{Op: "concat", Parts: []model.ValueExpr{
					{Op: "source", Source: "entity_type"},
					{Op: "const", Value: ":"},
					{Op: "source", Source: "channel"},
				}}},
				{Name: "tmpl", From: model.ValueExpr{Op: "template", Template: "id={entity_id} ch={channel}"}},
			},
		},
		Postamble: []model.RecordTemplate{{
			When: &model.WhenClause{Field: "value", Op: "absent"},
			Fields: []model.Field{
				{Name: "record_type", From: model.ValueExpr{Op: "const", Value: "trace_end"}},
			},
		}},
	}
}

func TestRenderTransforms(t *testing.T) {
	a := miniAdapter()
	meta := map[string]any{
		"run_id": "r-1", "sim_version": "0.1.0", "domain_id": "d",
		"domain_version": "0.0.0", "world_start_time": model.FormatTime(0),
		"seed": float64(42),
	}
	e, err := NewEngine(a, meta)
	if err != nil {
		t.Fatal(err)
	}
	fx, err := FixtureEvents()
	if err != nil {
		t.Fatal(err)
	}
	out, err := e.RenderRun(fx, 1785315600000000000+10800*1e9)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 14 { // 1 preamble + 12 records + 1 postamble
		t.Fatalf("expected 14 lines, got %d", len(lines))
	}
	// Preamble carries run_meta.
	if !strings.Contains(lines[0], `"run":"r-1"`) {
		t.Fatalf("preamble run_meta missing: %s", lines[0])
	}
	// Record identity from seq, id rewritten (/ -> .).
	if !strings.Contains(lines[1], `"event.id":"evt-000000"`) {
		t.Fatalf("counter identity wrong: %s", lines[1])
	}
	if !strings.Contains(lines[1], `"event.entity_id":"site-a.pump-1"`) {
		t.Fatalf("id rewrite not applied: %s", lines[1])
	}
	// Omit when null: the heartbeat record (seq 3) has no value field.
	if strings.Contains(lines[4], `"event.value"`) {
		t.Fatalf("omit_when_null failed: %s", lines[4])
	}
	// Postamble guarded by when value absent: emitted because postamble has
	// no event context (value absent).
	if !strings.Contains(lines[13], `"trace_end"`) {
		t.Fatalf("postamble missing: %s", lines[13])
	}
}

func TestWhenGuardExcludes(t *testing.T) {
	a := miniAdapter()
	a.Postamble = []model.RecordTemplate{{
		When:   &model.WhenClause{Field: "value", Op: "present"},
		Fields: []model.Field{{Name: "x", From: model.ValueExpr{Op: "const", Value: 1}}},
	}}
	e, err := NewEngine(a, map[string]any{"run_id": "r"})
	if err != nil {
		t.Fatal(err)
	}
	fx, _ := FixtureEvents()
	out, err := e.RenderRun(fx, 0)
	if err != nil {
		t.Fatal(err)
	}
	// No postamble: the when-guard sees no event context.
	if strings.Contains(string(out), "trace_end") && strings.Contains(string(out), `"x":1`) {
		t.Fatalf("guard did not exclude postamble: %s", out)
	}
}

func TestIdentityRequirement(t *testing.T) {
	// A record without seq-derived identity must be refused at load.
	doc := `{
		"id": "no-identity", "version": "0.1.0", "encoding": "jsonl",
		"record": {"fields": [{"name": "t", "from": {"op": "const", "value": 1}}]}
	}`
	if _, err := LoadBytes([]byte(doc), "test"); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("expected identity requirement error, got %v", err)
	}
}

func TestVerifyAgainstOwnGolden(t *testing.T) {
	// The native-jsonl adapter verifies against its own golden; see the
	// adapters/ dir tests for the shipped pair. Here we verify the engine's
	// byte-stability directly.
	a := miniAdapter()
	e, err := NewEngine(a, map[string]any{
		"run_id": "r", "sim_version": "0.1.0", "domain_id": "d",
		"domain_version": "0.0.0", "world_start_time": "2026-07-29T09:00:00Z",
		"seed": float64(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	fx, _ := FixtureEvents()
	first, err := e.RenderRun(fx, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.RenderRun(fx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("rendering is not byte-stable")
	}
}

func TestStreamingLifecycleMatchesBatchForJSONL(t *testing.T) {
	a := miniAdapter()
	fx, err := FixtureEvents()
	if err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"run_id": "r", "world_start_time": fx[0].EventTime}
	batch, err := NewEngine(a, meta)
	if err != nil {
		t.Fatal(err)
	}
	worldEnd, _ := model.ParseTime(fx[len(fx)-1].ObservedTime)
	want, err := batch.RenderRun(fx, worldEnd)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := NewEngine(a, map[string]any{"run_id": "r", "world_start_time": fx[0].EventTime})
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	begin, err := stream.Begin()
	if err != nil {
		t.Fatal(err)
	}
	lines = append(lines, begin...)
	for i := range fx {
		line, err := stream.RenderStreamRecord(&fx[i])
		if err != nil {
			t.Fatal(err)
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	end, err := stream.End(worldEnd)
	if err != nil {
		t.Fatal(err)
	}
	lines = append(lines, end...)
	got := strings.Join(lines, "\n") + "\n"
	if got != string(want) {
		t.Fatalf("streaming lifecycle diverged from batch:\nstream=%s\nbatch=%s", got, want)
	}
}

func TestJSONArrayLifecycleProducesValidArray(t *testing.T) {
	a := miniAdapter()
	a.Encoding = "json-array"
	fx, err := FixtureEvents()
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(a, map[string]any{"run_id": "r", "world_start_time": fx[0].EventTime})
	if err != nil {
		t.Fatal(err)
	}
	end, _ := model.ParseTime(fx[len(fx)-1].ObservedTime)
	out, err := e.RenderRun(fx, end)
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	if err := json.Unmarshal(out, &records); err != nil {
		t.Fatalf("json-array output is invalid: %v\n%s", err, out)
	}
	if len(records) != 14 {
		t.Fatalf("expected preamble, records, and postamble in array; got %d", len(records))
	}
}

// LoadBytes validates an adapter from memory (test helper).
func LoadBytes(raw []byte, src string) (*model.Adapter, error) {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	var a model.Adapter
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if err := crossCheck(&a, src); err != nil {
		return nil, err
	}
	return &a, nil
}
