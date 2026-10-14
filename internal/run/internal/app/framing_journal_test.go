package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheObserverJournalShowsTheTraceRecordForRecordIncludingTheAdapterFraming(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	framed, err := adapter.Load(testsupport.Adapter("agentic-stream"))
	if err != nil {
		t.Fatal(err)
	}
	dir, start := t.TempDir(), model.DefaultStartTimeNS+4*3600*1e9
	r, err := New(context.Background(), Config{Domain: spec, Adapter: framed, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+3600*1e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	records, err := decodeJournalBytes(readFile(t, filepath.Join(dir, journal.ObserverFile)))
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	var kinds []any
	for _, record := range records {
		if record.Type == journal.TypeObservation {
			lines, kinds = append(lines, record.Fields["line"].(string)), append(kinds, record.Fields["framing"])
		}
	}
	if got, want := strings.Join(lines, "\n"), strings.TrimSpace(string(r.Trace())); got != want {
		t.Fatalf("observer journal observations differ from the trace:\n%s\n--- trace ---\n%s", got, want)
	}
	if kinds[0] != journal.FramingPreamble || kinds[len(kinds)-1] != journal.FramingPostamble {
		t.Fatalf("framing marks = %v, want the preamble first and the postamble last", kinds)
	}
}

func TestAnAdapterWithoutFramingJournalsNoFramingRecords(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := pondConfig(t)
	cfg.JournalDir = dir
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, cfg.StartTimeNS+3600*1e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile)); err != nil || strings.Contains(string(raw), `"framing"`) {
		t.Fatalf("a native trace has no preamble or postamble: err %v", err)
	}
}
