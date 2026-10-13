package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	pinnedRunID  = "r-3h6wx9adsnlft"
	pinnedDomain = "sha256:dfb37e19c2c3137194d23652a5531a07605fb909bce59e2f63724191545adbe1"
)

var truthWords = []string{"aerator_failure", "duplicate_burst", "dropped_by_perturbation", "silent_no_effect", "\"seed\"", "Seed"}

func rewrite(t *testing.T, dir, name string, edit func([]byte) []byte) {
	t.Helper()
	path := filepath.Join(dir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, edit(raw), 0o600); err != nil {
		t.Fatal(err)
	}
}

func replaceIn(old, replacement string) func([]byte) []byte {
	return func(raw []byte) []byte { return bytes.ReplaceAll(raw, []byte(old), []byte(replacement)) }
}

func TestReportShowsNoTruthFromARunJsonThatBelongsToAnotherRun(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, old, other, differs string }{
		{"run id", `"run_id": "` + pinnedRunID + `"`, `"run_id": "r-other"`, "run_id"},
		{"sim version", `"sim_version": "0.3.0"`, `"sim_version": "9.9.9"`, "sim_version"},
		{"domain", pinnedDomain, "sha256:" + strings.Repeat("0", 64), "domain"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := copyFolder(t, testsupport.ReportRun())
			writeLabel(t, dir, 0)
			rewrite(t, dir, "run.json", replaceIn(tc.old, tc.other))
			built, err := Build(t.Context(), dir)
			if err != nil {
				t.Fatal(err)
			}
			if built.Truth != nil || built.Run.Seed != nil || built.Scoring != nil || built.Digests != nil || built.Run.Ended {
				t.Fatalf("a foreign run.json leaked into the report: %+v", built)
			}
			if !strings.Contains(built.TruthUnavailable, tc.differs) || built.Run.RunID != pinnedRunID {
				t.Fatalf("truth_unavailable %q, run id %q: want the reason naming %s and the journals' identity", built.TruthUnavailable, built.Run.RunID, tc.differs)
			}
			for _, format := range []string{JSONFormat, Markdown} {
				text := string(render(t, dir, format))
				for _, word := range truthWords {
					if strings.Contains(text, word) {
						t.Errorf("the %s report leaks %q", format, word)
					}
				}
			}
			if !strings.Contains(string(render(t, dir, Markdown)), "Truth unavailable: run.json belongs to another run") {
				t.Error("markdown does not say why the truth is unavailable")
			}
		})
	}
}

func TestReportShowsNoTruthFromASealedJournalThatBelongsToAnotherRun(t *testing.T) {
	t.Parallel()
	for name, edit := range map[string]func([]byte) []byte{
		"run id": replaceIn(pinnedRunID, "r-other"),
		"seed":   replaceIn(`"seed":7`, `"seed":8`),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := copyFolder(t, testsupport.ReportRun())
			rewrite(t, dir, "sealed/journal.jsonl", edit)
			built, err := Build(t.Context(), dir)
			if err != nil {
				t.Fatal(err)
			}
			if built.Truth != nil || built.Counts.LedgerRows != nil || !strings.Contains(built.TruthUnavailable, "sealed/journal.jsonl belongs to another run") {
				t.Fatalf("a foreign sealed journal: truth %v, ledger rows %v, reason %q", built.Truth, built.Counts.LedgerRows, built.TruthUnavailable)
			}
		})
	}
}

func TestReportOfARunWhoseFilesAgreeShowsItsTruth(t *testing.T) {
	t.Parallel()
	built, err := Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	if built.Truth == nil || built.TruthUnavailable != "" || len(built.TornTail) != 0 {
		t.Fatalf("a consistent folder: truth %v, unavailable %q, torn %v", built.Truth, built.TruthUnavailable, built.TornTail)
	}
}

func cutLastLine(raw []byte) []byte {
	return raw[:len(raw)-20]
}

func TestReportDropsATornLastLineAndSaysSo(t *testing.T) {
	t.Parallel()
	whole, err := Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	dir := copyFolder(t, testsupport.ReportRun())
	rewrite(t, dir, "ledger.jsonl", cutLastLine)
	rewrite(t, dir, "journal.jsonl", cutLastLine)
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatalf("a torn tail must not fail the report: %v", err)
	}
	if strings.Join(built.TornTail, ",") != "journal.jsonl,ledger.jsonl" || *built.Counts.LedgerRows != *whole.Counts.LedgerRows-1 {
		t.Fatalf("torn tail %v, ledger rows %d (whole run %d)", built.TornTail, *built.Counts.LedgerRows, *whole.Counts.LedgerRows)
	}
	text := string(render(t, dir, Markdown))
	if !strings.Contains(text, "## Torn tail") || !strings.Contains(text, "ledger.jsonl: the last line was cut short and dropped.") {
		t.Fatalf("markdown does not report the torn tail:\n%s", text)
	}
}

func TestReportOfACrashedExploreRunToleratesATornLedger(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	for _, name := range []string{"run.json", "verdict.json"} {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	rewrite(t, dir, "journal.jsonl", replaceIn(`"mode":"evaluate"`, `"mode":"explore"`))
	rewrite(t, dir, "ledger.jsonl", cutLastLine)
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatalf("the crashed run's report: %v", err)
	}
	if built.Run.Ended || built.Truth == nil || len(built.TornTail) == 0 {
		t.Fatalf("crashed explore run: ended %v, truth %v, torn %v", built.Run.Ended, built.Truth, built.TornTail)
	}
}

func TestReportStillFailsOnACorruptLineInTheMiddleOfTheLedger(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	rewrite(t, dir, "ledger.jsonl", func(raw []byte) []byte {
		return append([]byte("{not json\n"), raw...)
	})
	if _, err := Build(t.Context(), dir); err == nil || !strings.Contains(err.Error(), "ledger.jsonl line 1") {
		t.Fatalf("err = %v, want one naming ledger.jsonl line 1", err)
	}
}
