package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestOriginAndBindingRecordsReachTheObserverJournalAndLeaveTheReplayUnchanged(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	r.JournalOrigin(map[string]any{"instance": "zone-01", "device": "zone-01", "board": "uno-07", "environment": "hardware", "boot_id": "61"})
	r.JournalBinding(map[string]any{"outcome": "matched", "issues": []any{}})
	r.JournalSession(map[string]any{"mode":"connected","plan":map[string]any{"domain":"aquaculture-pond"}})
	r.JournalLive(map[string]any{"host":"online","gateway":"not_applicable","devices":[]any{},"commands":[]any{},"marks":[]any{},"commands_truncated":false})
	if err := r.FlushJournal(); err != nil {
		t.Fatal(err)
	}
	if early, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile)); err != nil || !strings.Contains(string(early), `"type":"binding"`) {
		t.Fatalf("observer journal before the end = %q, %v, want the binding on disk once flushed", early, err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, journal.ObserverFile))
	if err != nil || !strings.Contains(string(raw), `"type":"origin"`) || !strings.Contains(string(raw), `"type":"binding"`) {
		t.Fatalf("observer journal = %q, %v, want the origin and the binding records", raw, err)
	}
	if !strings.Contains(string(raw), `"type":"session"`) || !strings.Contains(string(raw), `"type":"live"`) {
		t.Fatalf("new session metadata and live evidence missing: %s",raw)
	}
	dirs := run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs); err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want the session records outside what a replay reproduces", res, err)
	}
}
