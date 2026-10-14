package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func incompleteRunFolder(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 12, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	r.Sink = failingSink{}
	if _, err := r.Advance(t.Context(), start+3600*1e9, false); err == nil || !strings.Contains(err.Error(), "injected sink failure") {
		t.Fatalf("the failing sink must abort the advance, got %v", err)
	}
	if _, err := r.End(dir); err == nil || !strings.Contains(err.Error(), "injected sink failure") {
		t.Fatalf("End must return the run failure, got %v", err)
	}
	return dir
}

func verifyFolder(t *testing.T, dir string) *ReplayResult {
	t.Helper()
	dirs := Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	res, err := ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestAGenuinelyIncompleteRunIsNeverAMatchAndSaysWhy(t *testing.T) {
	t.Parallel()
	dir := incompleteRunFolder(t)
	res := verifyFolder(t, dir)
	if res.Matches || res.Reason != ReasonIncomplete || !res.Incomplete || res.File != "" {
		t.Fatalf("%+v; want matches:false, reason incomplete, incomplete:true", res)
	}
	if !strings.Contains(res.Detail, "injected sink failure") || !strings.Contains(res.Detail, "cannot be verified as complete") {
		t.Fatalf("detail %q; want the run failure and the rule", res.Detail)
	}
	if !Reproduced(res) {
		t.Fatal("what an incomplete run did record still replays; grading may use it")
	}
}

func TestAnIncompleteRunIsHeldToTheEvidenceItRecorded(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		tamper       func(t *testing.T, dir string)
		reason, file string
		stamped      bool
	}{
		{"a ledger row appended", func(t *testing.T, dir string) { appendTo(t, filepath.Join(dir, model.FileLedger), "{}\n") },
			ReasonFileDiffers, model.FileLedger, false},
		{"the history deleted", func(t *testing.T, dir string) { remove(t, filepath.Join(dir, model.FileHistory)) },
			ReasonFileMissing, model.FileHistory, false},
		{"the sealed journal deleted", func(t *testing.T, dir string) { remove(t, filepath.Join(dir, model.FileDirectorJournal)) },
			ReasonFileMissing, model.FileDirectorJournal, false},
		{"the stamp cleared on an incomplete run", func(t *testing.T, dir string) {
			editRunJSON(t, filepath.Join(dir, model.FileRunArtifact), func(doc map[string]any) { delete(doc, "incomplete") })
		}, ReasonJournalDivergence, "", true},
		{"recorded as reproducible", func(t *testing.T, dir string) {
			editRunJSON(t, filepath.Join(dir, model.FileRunArtifact), func(doc map[string]any) { doc["reproducible"] = true })
		}, ReasonIncompleteStamp, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := incompleteRunFolder(t)
			tc.tamper(t, dir)
			res := verifyFolder(t, dir)
			if res.Matches || res.Reason != tc.reason || res.File != tc.file || (tc.stamped && !strings.Contains(res.Detail, "closed with incomplete=true")) {
				t.Fatalf("%+v; want matches:false, reason %s, file %q", res, tc.reason, tc.file)
			}
		})
	}
}

func TestAnIncompleteRunMissingOptionalEvidenceItNeverRecordedStillReportsIncomplete(t *testing.T) {
	t.Parallel()
	dir := incompleteRunFolder(t)
	editRunJSON(t, filepath.Join(dir, model.FileRunArtifact), func(doc map[string]any) {
		delete(doc, "history_digest")
		delete(doc, "ledger_digest")
	})
	remove(t, filepath.Join(dir, model.FileHistory))
	res := verifyFolder(t, dir)
	if res.Matches || res.Reason != ReasonIncomplete {
		t.Fatalf("%+v; want matches:false, reason incomplete", res)
	}
}

func appendTo(t *testing.T, path, text string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(text); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
