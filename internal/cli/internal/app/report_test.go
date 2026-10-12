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
		{"markdown without a file", []string{"report", "--format", "markdown", dir}},
		{"unknown format", []string{"report", "--format", "html", "--out", filepath.Join(t.TempDir(), "r"), dir}},
		{"no run folder", []string{"report"}},
		{"two run folders", []string{"report", dir, dir}},
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
