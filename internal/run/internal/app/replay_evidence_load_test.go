package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func editRunJSON(t *testing.T, path string, edit func(doc map[string]any)) {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(readFile(t, path), &doc); err != nil {
		t.Fatal(err)
	}
	edit(doc)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestVerdictLoadingToleratesOnlyAbsentEvidenceDigests(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		edit         func(doc map[string]any)
		strictFails  bool
		verdictFails bool
	}{
		{"untouched", func(map[string]any) {}, false, false},
		{"a ledger digest removed", func(doc map[string]any) { delete(doc, "ledger_digest") }, true, false},
		{"every evidence digest removed", func(doc map[string]any) {
			for _, field := range []string{"expected_trace_digest", "journal_digest", "sealed_journal_digest", "journal_file_digest",
				"sealed_journal_file_digest", "ledger_digest", "history_digest"} {
				delete(doc, field)
			}
		}, true, false},
		{"a digest removed and another property broken", func(doc map[string]any) {
			delete(doc, "ledger_digest")
			doc["seed"] = "seven"
		}, true, true},
		{"a digest malformed", func(doc map[string]any) { doc["ledger_digest"] = "sha256:00" }, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			scriptedJournalRun(t, dir, nil)
			path := filepath.Join(dir, "run.json")
			editRunJSON(t, path, tc.edit)
			if _, err := LoadArtifact(path); (err != nil) != tc.strictFails {
				t.Fatalf("LoadArtifact error = %v, want failure %v", err, tc.strictFails)
			}
			if _, err := loadArtifactForVerdict(path); (err != nil) != tc.verdictFails {
				t.Fatalf("loadArtifactForVerdict error = %v, want failure %v", err, tc.verdictFails)
			}
		})
	}
}

func TestARunWhoseTraceSinkIsOutsideTheFolderStillKeepsItsTraceInTheFolder(t *testing.T) {
	t.Parallel()
	dir, elsewhere := t.TempDir(), filepath.Join(t.TempDir(), "trace.jsonl")
	scriptedJournalRun(t, dir, func(cfg *Config) { cfg.SinkTarget = elsewhere })
	if string(readFile(t, filepath.Join(dir, "trace.jsonl"))) != string(readFile(t, elsewhere)) {
		t.Fatal("the folder's trace.jsonl must equal the trace written to the sink target")
	}
	dirs := Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	res, err := ReplayFile(t.Context(), filepath.Join(dir, "run.json"), dirs)
	if err != nil || !res.Matches || res.Reason != ReasonMatch {
		t.Fatalf("replay = %+v, %v; want a match", res, err)
	}
}
