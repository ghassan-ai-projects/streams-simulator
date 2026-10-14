package run_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

var evidenceDigestFields = []struct {
	file   string
	fields []string
}{
	{model.FileTrace, []string{"expected_trace_digest"}},
	{model.FileLedger, []string{"ledger_digest"}},
	{model.FileHistory, []string{"history_digest"}},
	{model.FileObserverJournal, []string{"journal_digest", "journal_file_digest"}},
	{model.FileDirectorJournal, []string{"sealed_journal_digest", "sealed_journal_file_digest"}},
}

func deleteSecondRow(raw []byte) []byte {
	rows := bytes.SplitAfter(raw, []byte("\n"))
	return bytes.Join(append(rows[:1:1], rows[2:]...), nil)
}

func dropDigests(fields []string) func(map[string]any) {
	return func(doc map[string]any) {
		for _, field := range fields {
			delete(doc, field)
		}
	}
}

func TestVerifyNamesWhyAClosedRunWithTamperedEvidenceDoesNotMatch(t *testing.T) {
	t.Parallel()
	kinds := []struct {
		name   string
		tamper func(t *testing.T, dir, file string, fields []string)
		reason string
	}{
		{"a row deleted with the digest kept", func(t *testing.T, dir, file string, _ []string) {
			rewrite(t, filepath.Join(dir, file), deleteSecondRow)
		}, "file_differs"},
		{"a row deleted with the digest removed", func(t *testing.T, dir, file string, fields []string) {
			rewrite(t, filepath.Join(dir, file), deleteSecondRow)
			editArtifact(t, filepath.Join(dir, model.FileRunArtifact), dropDigests(fields))
		}, "digest_missing"},
		{"the digest removed with the file intact", func(t *testing.T, dir, _ string, fields []string) {
			editArtifact(t, filepath.Join(dir, model.FileRunArtifact), dropDigests(fields))
		}, "digest_missing"},
		{"the file deleted", func(t *testing.T, dir, file string, _ []string) {
			if err := os.Remove(filepath.Join(dir, file)); err != nil {
				t.Fatal(err)
			}
		}, "file_missing"},
	}
	for _, evidence := range evidenceDigestFields {
		for _, kind := range kinds {
			t.Run(evidence.file+"/"+kind.name, func(t *testing.T) {
				t.Parallel()
				dir, dirs := publishedRun(t)
				kind.tamper(t, dir, evidence.file, evidence.fields)
				res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs)
				if err != nil || res.Matches || res.Reason != kind.reason || res.File != evidence.file || !strings.Contains(res.Detail, evidence.file) {
					t.Fatalf("%+v, %v; want matches:false, reason %s, file %s, a detail naming it", res, err, kind.reason, evidence.file)
				}
			})
		}
	}
}

func TestVerifyReportsMatchForACleanClosedRun(t *testing.T) {
	t.Parallel()
	dir, dirs := publishedRun(t)
	res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs)
	if err != nil || !res.Matches || res.Reason != "match" || res.File != "" || res.Detail != "" {
		t.Fatalf("%+v, %v; want a match with reason match and no file or detail", res, err)
	}
}

func TestVerifyNamesTheRecordedDigestThatTheReplayDoesNotReproduce(t *testing.T) {
	t.Parallel()
	zero := "sha256:" + strings.Repeat("0", 64)
	cases := map[string]struct {
		field, reason, file string
	}{
		"the trace digest":   {"expected_trace_digest", "trace_divergence", ""},
		"the ledger digest":  {"ledger_digest", "evidence_divergence", model.FileLedger},
		"the history digest": {"history_digest", "evidence_divergence", model.FileHistory},
		"the journal digest": {"journal_digest", "journal_divergence", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir, dirs := publishedRun(t)
			editArtifact(t, filepath.Join(dir, model.FileRunArtifact), func(doc map[string]any) { doc[tc.field] = zero })
			res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs)
			if err != nil || res.Matches || res.Reason != tc.reason || res.File != tc.file {
				t.Fatalf("%+v, %v; want matches:false, reason %s, file %q", res, err, tc.reason, tc.file)
			}
		})
	}
}

func TestAnIncompleteRunKeepsItsOptionalEvidenceDigests(t *testing.T) {
	t.Parallel()
	dir, dirs := publishedRun(t)
	fields := []string{"ledger_digest", "history_digest", "journal_file_digest", "sealed_journal_file_digest"}
	editArtifact(t, filepath.Join(dir, model.FileRunArtifact), func(doc map[string]any) {
		doc["incomplete"] = true
		dropDigests(fields)(doc)
	})
	if err := os.Remove(filepath.Join(dir, model.FileHistory)); err != nil {
		t.Fatal(err)
	}
	res, err := run.ReplayFile(t.Context(), filepath.Join(dir, model.FileRunArtifact), dirs)
	if err != nil || res.Reason == "digest_missing" || res.Reason == "file_missing" {
		t.Fatalf("%+v, %v; an incomplete run is not held to the closed-run evidence rule", res, err)
	}
}
