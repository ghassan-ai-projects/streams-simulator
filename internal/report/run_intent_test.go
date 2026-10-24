package report_test

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func TestTheReportCarriesTheRequestedEndAndTheTerminationTheRunRecorded(t *testing.T) {
	t.Parallel()
	end := model.DefaultStartTimeNS + 3600e9
	dir := folderWith(t, func(_ string, r *run.Run) {
		if err := r.SetPlan(model.RunPlan{RequestedEndNS: end, Rate: 60}); err != nil {
			t.Fatal(err)
		}
		if err := r.SetTermination(model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: model.DefaultStartTimeNS + 120e9}); err != nil {
			t.Fatal(err)
		}
	})
	built, err := report.Build(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if built.Run.RequestedEndNS != end || built.Run.Termination == nil || built.Run.Termination.Reason != model.TerminationStoppedByOperator ||
		built.Run.Termination.AtNS != model.DefaultStartTimeNS+120e9 {
		t.Fatalf("run = %+v, want the requested end and the operator stop", built.Run)
	}
	assertReportValid(t, built, `"requested_end_ns"`, `"rate": 60`, `"termination"`, `"stopped_by_operator"`)
	markdown, err := report.Render(built, report.FormatMarkdown)
	if err != nil || !strings.Contains(string(markdown), "Stopped by operator") || !strings.Contains(string(markdown), "60 simulated s per wall s") {
		t.Fatalf("markdown = %q, %v, want the stop named", markdown, err)
	}
}

func TestTheReportOfARunWithoutIntentHasNeitherField(t *testing.T) {
	t.Parallel()
	built, err := report.Build(t.Context(), folderWith(t, func(string, *run.Run) {}))
	if err != nil {
		t.Fatal(err)
	}
	if built.Run.RequestedEndNS != 0 || built.Run.Rate != 0 || built.Run.Termination != nil {
		t.Fatalf("run = %+v, want no intent", built.Run)
	}
	assertReportValid(t, built)
	raw, err := report.Render(built, report.FormatJSON)
	if err != nil || strings.Contains(string(raw), "requested_end_ns") || strings.Contains(string(raw), `"rate"`) || strings.Contains(string(raw), "termination") {
		t.Fatalf("report = %s, %v, want neither field", raw, err)
	}
}

func assertReportValid(t *testing.T, built *report.Report, wanted ...string) {
	t.Helper()
	raw, err := report.Render(built, report.FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range wanted {
		if !strings.Contains(string(raw), text) {
			t.Errorf("report lacks %s:\n%s", text, raw)
		}
	}
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
}

func TestTheReportOfAFastPlanHasARequestedEndButNoRate(t *testing.T) {
	t.Parallel()
	end := model.DefaultStartTimeNS + 600e9
	built, err := report.Build(t.Context(), folderWith(t, func(_ string, r *run.Run) {
		if err := r.SetPlan(model.RunPlan{RequestedEndNS: end}); err != nil {
			t.Fatal(err)
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	assertReportValid(t, built, `"requested_end_ns"`)
	raw, err := report.Render(built, report.FormatJSON)
	if err != nil || built.Run.Rate != 0 || strings.Contains(string(raw), `"rate"`) {
		t.Fatalf("report = %s, %v, want no rate for a plan run as fast as possible", raw, err)
	}
}
