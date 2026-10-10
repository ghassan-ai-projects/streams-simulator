package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// buildArtifact runs the config and returns its artifact.
func buildArtifact(t *testing.T, cfg Config) *model.RunArtifact {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), cfg.StartTimeNS+6*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

// TestRunSinkEquivalence: inproc and file produce byte-identical traces.
func TestRunSinkEquivalence(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	run := func(sink, target string) []byte {
		r, err := New(context.Background(), Config{
			Domain: spec, Adapter: a, Seed: 7, SinkName: sink,
			SinkTarget: target, TimeMode: model.TimeStepped, StartTimeNS: start,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Advance(context.Background(), start+2*3600*1e9, false); err != nil {
			t.Fatal(err)
		}
		art, err := r.End("")
		if err != nil {
			t.Fatal(err)
		}
		return []byte(art.ExpectedTraceDigest + "|" + string(r.trace))
	}
	inproc := run(model.SinkInproc, "")
	file := run(model.SinkFile, filepath.Join(t.TempDir(), "trace.jsonl"))
	if string(inproc) != string(file) {
		t.Fatalf("sink equivalence broken: inproc and file diverged")
	}
}

func TestStreamingRunIncludesAdapterPreambleAndPostamble(t *testing.T) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(agenticAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(r.Trace())), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected preamble, event, and postamble; got %d lines", len(lines))
	}
	if !strings.Contains(lines[0], `"record_type":"runtime_config"`) {
		t.Fatalf("missing runtime preamble: %s", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], `"record_type":"trace_end"`) {
		t.Fatalf("missing trace postamble: %s", lines[len(lines)-1])
	}
}

func TestGeneratedTraceTrailerFollowsEventArrivals(t *testing.T) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(agenticAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
		EntityIDs: []string{"site-a/pond-1"}, Noiseless: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+60*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	if !art.Reproducible {
		t.Fatal("generated trace must remain reproducible")
	}

	var maxArrival int64
	var until int64
	haveEvent, haveTrailer := false, false
	for i, line := range strings.Split(strings.TrimSpace(string(r.Trace())), "\n") {
		var record struct {
			RecordType string `json:"record_type"`
			Event      struct {
				ArrivalTime string `json:"arrival_time"`
			} `json:"event"`
			Until string `json:"until"`
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode trace line %d: %v", i, err)
		}
		switch record.RecordType {
		case "event":
			arrival, err := model.ParseTime(record.Event.ArrivalTime)
			if err != nil {
				t.Fatalf("parse event arrival on line %d: %v", i, err)
			}
			if !haveEvent || arrival > maxArrival {
				maxArrival = arrival
			}
			haveEvent = true
		case "trace_end":
			var err error
			until, err = model.ParseTime(record.Until)
			if err != nil {
				t.Fatalf("parse trailer until on line %d: %v", i, err)
			}
			haveTrailer = true
		}
	}
	if !haveEvent || !haveTrailer {
		t.Fatalf("trace must contain event(s) and a trace_end trailer")
	}
	if maxArrival >= until {
		t.Fatalf("trace_end.until %s must be later than last arrival %s", model.FormatTime(until), model.FormatTime(maxArrival))
	}
}

func TestSinkFailureMarksRunIncompleteWithoutPanic(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 12, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Sink = failingSink{}
	if _, err := r.Advance(context.Background(), start+3600*1e9, false); err == nil {
		t.Fatal("sink failure must be returned from Advance")
	}
	art, endErr := r.End("")
	if endErr == nil {
		t.Fatal("incomplete run must retain its failure")
	}
	if art == nil || !art.Incomplete || art.Error == "" {
		t.Fatalf("incomplete artifact missing failure state: %+v", art)
	}
}
