package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const withheldLine = "Truth withheld: the run has not ended (no run.json)."

func render(t *testing.T, dir, format string) []byte {
	t.Helper()
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatalf("build %s: %v", dir, err)
	}
	out, err := Render(built, format)
	if err != nil {
		t.Fatalf("render %s: %v", format, err)
	}
	return out
}

func copyFolder(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestReportOfThePinnedRunMatchesItsGoldenDocuments(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ format, golden string }{{JSONFormat, "report.golden.json"}, {Markdown, "report.golden.md"}} {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()
			got := render(t, testsupport.ReportRun(), tc.format)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				if err := os.WriteFile(testsupport.ReportGolden(tc.golden), got, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(testsupport.ReportGolden(tc.golden))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs from %s:\n%s", tc.format, tc.golden, got)
			}
		})
	}
}

func TestReportIsTheSameBytesForTheSameFolderWhereverItLives(t *testing.T) {
	t.Parallel()
	for _, format := range []string{JSONFormat, Markdown} {
		first := render(t, testsupport.ReportRun(), format)
		second := render(t, copyFolder(t, testsupport.ReportRun()), format)
		if !bytes.Equal(first, second) {
			t.Fatalf("%s report differs between two copies of one folder", format)
		}
	}
}

func TestReportValidatesAgainstTheRunReportContract(t *testing.T) {
	t.Parallel()
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := model.DecodeBytes(render(t, testsupport.ReportRun(), JSONFormat), &doc); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
}

func TestReportOfARunCutBeforeItsArtifactShowsItIncompleteAndWithholdsTruth(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	if err := os.Remove(filepath.Join(dir, "run.json")); err != nil {
		t.Fatal(err)
	}
	got := render(t, dir, JSONFormat)
	var report domain.Report
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatal(err)
	}
	if report.Run.Complete || report.Run.Ended || report.Truth != nil || report.Run.Seed != nil || report.Verification.Status != domain.NotChecked {
		t.Fatalf("open run report = %+v, want incomplete, not ended, no truth, no seed, not checked", report)
	}
	if c := report.Counts; c.LedgerRows != nil || c.Delivered != nil || c.Dropped != nil || c.Emitted != nil || c.Observed == nil {
		t.Fatalf("counts = %+v, want only the observed count", c)
	}
	for _, format := range []string{JSONFormat, Markdown} {
		text := string(render(t, dir, format))
		for _, secret := range []string{"aerator_failure", "duplicate_burst", "dropped_by_perturbation", "silent_no_effect", "seed", "Seed",
			"ledger_rows", "Ledger rows", "Dropped", "delivered_by_reason"} {
			if strings.Contains(text, secret) {
				t.Errorf("the %s report of an open evaluate run leaks %q", format, secret)
			}
		}
	}
	if !strings.Contains(string(render(t, dir, Markdown)), withheldLine) {
		t.Errorf("markdown does not say the truth is withheld")
	}
}

func TestReportShowsAnEditedTraceAsFilesDiffer(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	if err := os.WriteFile(filepath.Join(dir, "trace.jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if built.Verification.Status != domain.FilesDiffer || built.Verification.Checks[0].Status != domain.CheckDiffers {
		t.Fatalf("verification = %+v, want files_differ with a trace mismatch", built.Verification)
	}
}

func TestReportOfACopiedFolderWhoseScoringFailsIsTheSameBytes(t *testing.T) {
	t.Parallel()
	first, second := copyFolder(t, testsupport.ReportRun()), copyFolder(t, testsupport.ReportRun())
	for _, dir := range []string{first, second} {
		writeLabel(t, dir, 0)
	}
	one, other := render(t, first, JSONFormat), render(t, second, JSONFormat)
	if !bytes.Equal(one, other) || !strings.Contains(string(one), `"status": "failed"`) || strings.Contains(string(one), first) {
		t.Fatalf("failed scoring differs between copies or names the folder:\n%s\n%s", one, other)
	}
}

func TestBuildNamesAFolderWithNoRunEvidence(t *testing.T) {
	t.Parallel()
	_, err := Build(t.Context(), t.TempDir())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want a not-exist error", err)
	}
}

func TestBuildNamesAJournalThatDoesNotParse(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	if err := os.WriteFile(filepath.Join(dir, "journal.jsonl"), []byte("not json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(t.Context(), dir); err == nil || !strings.Contains(err.Error(), "journal.jsonl") {
		t.Fatalf("err = %v, want one naming journal.jsonl", err)
	}
}

func TestRenderRefusesAnUnknownFormat(t *testing.T) {
	t.Parallel()
	built, err := Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Render(built, "yaml"); !errors.Is(err, ErrUnknownFormat) {
		t.Fatalf("err = %v, want ErrUnknownFormat", err)
	}
}

func TestBuildNamesAVerdictThatBreaksItsContract(t *testing.T) {
	t.Parallel()
	dir := copyFolder(t, testsupport.ReportRun())
	if err := os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(`{"schema_version":"0.1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Build(t.Context(), dir)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractConsumerVerdict {
		t.Fatalf("err = %v, want a consumer-verdict contract violation", err)
	}
}
