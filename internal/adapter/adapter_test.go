package adapter_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func shippedAdapter(t *testing.T, id string) *model.Adapter {
	t.Helper()
	a, err := adapter.Load(filepath.Join("..", "..", "adapters", id+".adapter.json"))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNewEngineRefusesAMissingAdapter(t *testing.T) {
	t.Parallel()
	if _, err := adapter.NewEngine(nil, nil); !errors.Is(err, adapter.ErrNoAdapter) {
		t.Fatalf("err = %v, want ErrNoAdapter", err)
	}
}

func TestLoadBytesValidatesADocumentAndNamesItsSource(t *testing.T) {
	t.Parallel()
	loaded := shippedAdapter(t, "native-jsonl")
	again, err := adapter.LoadBytes(loaded.Raw, "in-memory")
	if err != nil || again.ID != loaded.ID {
		t.Fatalf("round trip: %v, %+v", err, again)
	}
	if _, err := adapter.LoadBytes([]byte(`{"id":"x"}`), "broken.json"); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Fatalf("invalid document: %v", err)
	}
}

func TestEngineRendersAStreamingSessionTheSameAsAWholeRun(t *testing.T) {
	t.Parallel()
	events, err := adapter.FixtureEvents()
	if err != nil || len(events) == 0 {
		t.Fatalf("fixture: %d events (%v)", len(events), err)
	}
	meta := map[string]any{
		"run_id": "r", "sim_version": "0.1.0", "domain_id": "d", "domain_version": "0.0.0",
		"world_start_time": events[0].EventTime, "world_end_time": events[len(events)-1].EventTime, "seed": float64(0),
	}
	end := int64(1)
	whole, err := adapter.NewEngine(shippedAdapter(t, "native-jsonl"), meta)
	if err != nil {
		t.Fatal(err)
	}
	want, err := whole.RenderRun(events, end)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := adapter.NewEngine(shippedAdapter(t, "native-jsonl"), meta)
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	lines, err := stream.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		got.WriteString(line + "\n")
	}
	for i := range events {
		line, err := stream.RenderStreamRecord(&events[i])
		if err != nil {
			t.Fatal(err)
		}
		if line != "" {
			got.WriteString(line + "\n")
		}
	}
	tail, err := stream.End(end)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range tail {
		got.WriteString(line + "\n")
	}
	if got.String() != string(want) {
		t.Fatalf("streaming session differs from RenderRun:\n%q\n%q", got.String(), want)
	}
}
