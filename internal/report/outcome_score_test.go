package report_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func strandingRun(t *testing.T) string {
	t.Helper()
	spec, err := domain.Load(testsupport.BeltConveyorDomain("b"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 1, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("D-3", "conveyor_drive__vfd_fault", model.DefaultStartTimeNS+60e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+600e9, false); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAReportAskedForTheOutcomeOfADomainWithOutcomeWeightsCarriesTheOutcomeScore(t *testing.T) {
	t.Parallel()
	dir := strandingRun(t)
	if plain, err := build(t.Context(), dir); err != nil || plain.OutcomeScore != nil {
		t.Fatalf("a default report = %+v (%v), want no outcome score and no replay", plain.OutcomeScore, err)
	}
	document, err := report.Produce(t.Context(), report.Request{Dir: dir, Outcome: true})
	if err != nil {
		t.Fatal(err)
	}
	built := document.(*report.Report)
	scored := built.OutcomeScore
	if scored == nil || scored.Status != "scored" || scored.Result == nil || !scored.Result.Valid || len(scored.Result.Verdicts) != 2 {
		t.Fatalf("outcome score %+v, want a scored valid story with verdicts against doing nothing and b4", scored)
	}
	asJSON, err := report.Render(built, report.FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := json.Unmarshal(asJSON, &doc); err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
	asMarkdown, err := report.Render(built, report.FormatMarkdown)
	if err != nil || !strings.Contains(string(asMarkdown), "## Outcome score") || !strings.Contains(string(asMarkdown), "| `policy:b4` |") {
		t.Fatalf("markdown = %s (%v), want an outcome score section", asMarkdown, err)
	}
}

func TestAReportAskedForTheOutcomeOfADomainWithoutOutcomeWeightsHasNoOutcomeScore(t *testing.T) {
	t.Parallel()
	document, err := report.Produce(t.Context(), report.Request{Dir: testsupport.ReportRun(), Outcome: true})
	if built, _ := document.(*report.Report); err != nil || built == nil || built.OutcomeScore != nil {
		t.Fatalf("report %+v (%v), want no outcome score", document, err)
	}
}
