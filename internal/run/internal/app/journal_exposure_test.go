package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// A drop perturbation leaves gaps in the native emission numbers. The observer
// journal holds exactly what the sink received, in order, so it shows no gap the
// trace does not.
func TestObserverJournalHoldsExactlyTheDeliveredRecordsOfADroppingRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkFile,
		SinkTarget: filepath.Join(dir, "trace.jsonl"), JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 0.5}, 0, 0); err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, r.World.Clock()+1800e9)
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	var delivered []string
	for _, rec := range recordsOfType(readJournal(t, dir, journal.ObserverFile), "observation") {
		if _, native := rec.Fields["event"]; native {
			t.Fatal("an observation must not carry the native event")
		}
		delivered = append(delivered, rec.Fields["line"].(string))
	}
	trace := strings.Split(strings.TrimSuffix(string(readFile(t, filepath.Join(dir, "trace.jsonl"))), "\n"), "\n")
	if len(delivered) == 0 || strings.Join(delivered, "\n") != strings.Join(trace, "\n") {
		t.Fatalf("observer journal holds %d records, trace %d lines; they must be the same lines in the same order", len(delivered), len(trace))
	}
}

// flipUnblinded rewrites run.json with unblinded set to false and returns the path.
func flipUnblinded(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "run.json")
	var doc map[string]any
	if err := json.Unmarshal(readFile(t, path), &doc); err != nil {
		t.Fatal(err)
	}
	doc["unblinded"] = false
	raw, _ := json.Marshal(doc)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestClearingTheUnblindedFlagOfAnExploreRunFailsVerification(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scriptedJournalRun(t, dir, func(cfg *Config) { cfg.Explore = true })
	path := flipUnblinded(t, dir)
	res, err := ReplayFile(t.Context(), path, Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil || res.Matches || !strings.Contains(res.Detail, "unblinded=false") {
		t.Fatalf("replay = %+v, %v; want a mismatch naming the stamp", res, err)
	}
}

func TestClearingTheUnblindedFlagOfARevealedRunFailsVerification(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	r.Unblind()
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	res, err := ReplayFile(t.Context(), flipUnblinded(t, dir), Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil || res.Matches || !bytes.Contains([]byte(res.Detail), []byte("director journal stream")) {
		t.Fatalf("replay = %+v, %v; want the director stream named", res, err)
	}
}
