package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestReportPrintsTheRunReportAsTheOneStdoutDocument(t *testing.T) {
	t.Parallel()
	got := invoke(t, "report", testsupport.ReportRun()).mustSucceed(t)
	doc := got.document(t, "report")
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	var parsed any
	if err := model.DecodeBytes([]byte(got.stdout), &parsed); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(parsed); len(errs) > 0 || doc["schema_version"] != "0.1" {
		t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
	golden, err := os.ReadFile(testsupport.ReportGolden("report.golden.json"))
	if err != nil || got.stdout != string(golden) {
		t.Fatalf("stdout differs from the golden report (%v):\n%s", err, got.stdout)
	}
}

func TestReportWritesMarkdownAndJSONFilesAndNamesThem(t *testing.T) {
	t.Parallel()
	for format, prefix := range map[string]string{"markdown": "# Run report: ", "json": "{\n  \"schema_version\""} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			out := filepath.Join(t.TempDir(), "report.out")
			doc := invoke(t, "report", "--format", format, "--out", out, testsupport.ReportRun()).mustSucceed(t).document(t, "report_written")
			written, err := os.ReadFile(out)
			if err != nil || !strings.HasPrefix(string(written), prefix) || doc["report_written"] != out || doc["format"] != format {
				t.Fatalf("doc = %v, file = %q (%v), want the %s report written to %s", doc, written, err, format, out)
			}
		})
	}
}

func TestReportRefusesInvocationsItCannotRunAsWritten(t *testing.T) {
	t.Parallel()
	dir := testsupport.ReportRun()
	cases := []struct {
		name string
		args []string
	}{
		{"unknown format", []string{"report", "--format", "html", "--out", filepath.Join(t.TempDir(), "r"), dir}},
		{"no run folder", []string{"report"}},
		{"two run folders", []string{"report", dir, dir}},
		{"a reference without the other run", []string{"report", "compare", "--a", dir}},
		{"a run folder given to compare", []string{"report", "compare", "--a", dir, "--b", dir, dir}},
		{"a comparison to replay", []string{"report", "compare", "--a", dir, "--b", dir, "--replay"}},
		{"a comparison format it does not know", []string{"report", "compare", "--a", dir, "--b", dir, "--format", "html"}},
		{"a side given to a plain report", []string{"report", "--a", dir, dir}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if failure := got.failure(t); got.code != 2 || failure["code"] != "usage_invalid" {
				t.Fatalf("exit %d failure %v, want exit 2 usage_invalid", got.code, failure)
			}
		})
	}
}

func TestReportFailureClasses(t *testing.T) {
	t.Parallel()
	torn := t.TempDir()
	if err := os.WriteFile(filepath.Join(torn, "journal.jsonl"), []byte("not json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, dir, code string }{
		{"folder without evidence", t.TempDir(), "file_not_found"},
		{"unreadable journal", torn, "report_failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, "report", tc.dir)
			if failure := got.failure(t); got.code != 1 || failure["code"] != tc.code {
				t.Fatalf("exit %d failure %v, want exit 1 %s", got.code, failure, tc.code)
			}
		})
	}
}

func TestMarkdownWithoutAFileIsStillOneJSONDocumentCarryingTheMarkdown(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "report", "--format", "markdown", testsupport.ReportRun()).mustSucceed(t).document(t, "report_markdown")
	text, _ := doc["markdown"].(string)
	if doc["format"] != "markdown" || !strings.HasPrefix(text, "# Run report: ") || !strings.Contains(text, "## Open questions") {
		t.Fatalf("doc = %v, want the markdown report inside the one JSON document", doc)
	}
}

func TestReportReplayReproducesARealRunAndTellsWhenItCannot(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	inputs := []string{"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}
	replayed := invoke(t, append([]string{"report", "--replay"}, append(inputs, dir)...)...).mustSucceed(t).document(t, "report")
	if replay, _ := replayed["replay"].(map[string]any); replay["status"] != "reproduces" {
		t.Fatalf("replay = %v, want reproduces", replay)
	}
	plain := invoke(t, "report", dir).mustSucceed(t).document(t, "report")
	if replay, _ := plain["replay"].(map[string]any); replay["status"] != "not_checked" {
		t.Fatalf("replay = %v, want not_checked without --replay", replay)
	}
}

func TestReportCompareIsOneSchemaValidDocumentAndTheSameBytesEveryTime(t *testing.T) {
	t.Parallel()
	dir := testsupport.ReportRun()
	first := invoke(t, "report", "compare", "--a", dir, "--b", dir).mustSucceed(t)
	first.document(t, "run_comparison")
	second := invoke(t, "report", "compare", "--b", dir, "--a", dir).mustSucceed(t)
	schema, err := jsonschema.CompileJSON(schemas.RunComparison())
	if err != nil {
		t.Fatal(err)
	}
	var parsed any
	if err := model.DecodeBytes([]byte(first.stdout), &parsed); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(parsed); len(errs) > 0 || first.stdout != second.stdout {
		t.Fatalf("comparison invalid or not deterministic (%d): %s", len(errs), jsonschema.FormatErrors(errs))
	}
	if !strings.Contains(first.stdout, `"single_run_per_side"`) {
		t.Fatalf("comparison = %s, want the single run limit", first.stdout)
	}
}

func TestReportCompareWritesMarkdownToAFileAndPrintsItInTheDocumentWithoutOne(t *testing.T) {
	t.Parallel()
	dir := testsupport.ReportRun()
	out := filepath.Join(t.TempDir(), "compare.md")
	doc := invoke(t, "report", "compare", "--a", dir, "--b", dir, "--format", "markdown", "--out", out).mustSucceed(t).document(t, "report_written")
	written, err := os.ReadFile(out)
	if err != nil || !strings.HasPrefix(string(written), "# Run comparison: ") || doc["format"] != "markdown" {
		t.Fatalf("doc = %v, file = %q (%v), want the markdown comparison written", doc, written, err)
	}
	inline := invoke(t, "report", "compare", "--a", dir, "--b", dir, "--format", "markdown").mustSucceed(t).document(t, "report_markdown")
	if text, _ := inline["markdown"].(string); text != string(written) {
		t.Fatalf("inline markdown differs from the written file:\n%s", text)
	}
}

func TestReportCompareRefusesARunThatHasNotEndedWithRunNotEnded(t *testing.T) {
	t.Parallel()
	open := t.TempDir()
	if err := os.CopyFS(open, os.DirFS(testsupport.ReportRun())); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(open, "run.json")); err != nil {
		t.Fatal(err)
	}
	got := invoke(t, "report", "compare", "--a", testsupport.ReportRun(), "--b", open)
	failure := got.failure(t)
	if got.code != 1 || failure["code"] != "run_not_ended" || !strings.Contains(failure["message"].(string), "run B") {
		t.Fatalf("exit %d failure %v, want exit 1 run_not_ended naming run B", got.code, failure)
	}
	missing := invoke(t, "report", "compare", "--a", t.TempDir(), "--b", open)
	if failure := missing.failure(t); missing.code != 1 || failure["code"] != "file_not_found" {
		t.Fatalf("failure %v, want file_not_found for a folder that is no run", failure)
	}
}
