package report_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestBuildAndRenderTheFacadeCoverBothFormats(t *testing.T) {
	t.Parallel()
	built, err := report.Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	asJSON, err := report.Render(built, report.FormatJSON)
	if err != nil || !strings.HasPrefix(string(asJSON), "{\n  \"schema_version\": \"0.1\"") {
		t.Fatalf("json = %q (%v), want a run-report document", asJSON, err)
	}
	asMarkdown, err := report.Render(built, report.FormatMarkdown)
	if err != nil || !strings.HasPrefix(string(asMarkdown), "# Run report: ") {
		t.Fatalf("markdown = %q (%v), want a titled report", asMarkdown, err)
	}
}

func TestRenderNamesAFormatItDoesNotKnow(t *testing.T) {
	t.Parallel()
	built, err := report.Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := report.Render(built, "html"); !errors.Is(err, report.ErrUnknownFormat) {
		t.Fatalf("err = %v, want ErrUnknownFormat", err)
	}
}
