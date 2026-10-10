package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestStreamingLifecycleMatchesBatchForJSONL(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
