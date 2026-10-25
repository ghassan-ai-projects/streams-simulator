package report_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
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

type planned struct {
	domain      string
	seed        uint64
	fault       bool
	clearAtS    int64
	endS, stopS int64
}

func (p planned) name() string {
	if p.domain == "" {
		return "aquaculture-pond"
	}
	return p.domain
}

func plannedFolder(t *testing.T, p planned) string {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain(p.name()))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: p.seed, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	drive(t, r, p)
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func drive(t *testing.T, r *run.Run, p planned) {
	t.Helper()
	at := func(seconds int64) int64 { return model.DefaultStartTimeNS + seconds*1e9 }
	if p.fault {
		id, err := r.InjectFault("site-a/pond-1", "aerator_failure", at(60), nil)
		if err != nil {
			t.Fatal(err)
		}
		if p.clearAtS > 0 {
			if err := r.ClearFault(id, at(p.clearAtS)); err != nil {
				t.Fatal(err)
			}
		}
	}
	end := cmpOr(p.stopS, p.endS)
	if _, err := r.Advance(t.Context(), at(end), false); err != nil {
		t.Fatal(err)
	}
	if err := r.SetPlan(model.RunPlan{RequestedEndNS: at(p.endS)}); err != nil {
		t.Fatal(err)
	}
	if p.stopS > 0 {
		if err := r.SetTermination(model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: at(p.stopS)}); err != nil {
			t.Fatal(err)
		}
	}
}

func cmpOr(a, b int64) int64 {
	if a != 0 {
		return a
	}
	return b
}

func comparing(t *testing.T, a, b planned) *report.Comparison {
	t.Helper()
	got, err := compare(t.Context(), plannedFolder(t, a), plannedFolder(t, b))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func limitCodes(c *report.Comparison) []string {
	codes := []string{}
	for _, limit := range c.Limits {
		codes = append(codes, limit.Code+limit.Side)
	}
	return codes
}

func TestTwoRunsOfOnePlanAndSeedHaveNoPlanDifferenceAndAreTheSameBytes(t *testing.T) {
	t.Parallel()
	base := planned{seed: 4, fault: true, clearAtS: 90, endS: 120}
	first, second := plannedFolder(t, base), plannedFolder(t, base)
	got, err := compare(t.Context(), first, second)
	if err != nil {
		t.Fatal(err)
	}
	plan := got.Plan
	if len(plan.Faults.Identical) != 1 || len(plan.Faults.Different)+len(plan.Faults.OnlyInA)+len(plan.Faults.OnlyInB) != 0 {
		t.Fatalf("faults = %+v, want one identical fault", plan.Faults)
	}
	if codes := limitCodes(got); !slices.Equal(codes, []string{"kpi_truth_absenta", "kpi_truth_absentb", "single_run_per_side"}) {
		t.Fatalf("limits = %v, want the single run limit and the absent KPI truth of a domain that declares none", codes)
	}
	one, _ := report.Render(got, report.FormatJSON)
	again, err := compare(t.Context(), first, second)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := report.Render(again, report.FormatJSON)
	if string(one) != string(other) {
		t.Fatal("two comparisons of the same folders differ")
	}
	assertComparisonValid(t, one)
}

func assertComparisonValid(t *testing.T, raw []byte) {
	t.Helper()
	schema, err := jsonschema.CompileJSON(schemas.RunComparison())
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("comparison does not match run-comparison-v0.1:\n%s\n%s", jsonschema.FormatErrors(errs), raw)
	}
}

func TestTheSamePlanWithAnotherSeedNamesTheSeedAndNothingInThePlan(t *testing.T) {
	t.Parallel()
	got := comparing(t, planned{seed: 4, fault: true, endS: 120}, planned{seed: 5, fault: true, endS: 120})
	if !slices.Contains(limitCodes(got), "different_seed") || len(got.Plan.Faults.Identical) != 1 {
		t.Fatalf("limits = %v faults = %+v, want different_seed and an identical fault", limitCodes(got), got.Plan.Faults)
	}
}

func TestAFaultOnlyTheSecondRunInjectedIsOnlyInB(t *testing.T) {
	t.Parallel()
	got := comparing(t, planned{seed: 4, endS: 120}, planned{seed: 4, fault: true, endS: 120})
	if len(got.Plan.Faults.OnlyInB) != 1 || got.Plan.Faults.OnlyInB[0].Fault != "aerator_failure" || len(got.Plan.Faults.OnlyInA) != 0 {
		t.Fatalf("faults = %+v, want the aerator failure only in B", got.Plan.Faults)
	}
}

func TestAFaultClearedEarlierInTheSecondRunDiffersInItsClear(t *testing.T) {
	t.Parallel()
	got := comparing(t, planned{seed: 4, fault: true, clearAtS: 100, endS: 120}, planned{seed: 4, fault: true, clearAtS: 80, endS: 120})
	pairs := got.Plan.Faults.Different
	if len(pairs) != 1 || !slices.Equal(pairs[0].Differences, []string{"cleared_ns"}) || *pairs[0].B.ClearedNS >= *pairs[0].A.ClearedNS {
		t.Fatalf("faults = %+v, want one pair differing in cleared_ns with B cleared earlier", got.Plan.Faults)
	}
}

func TestARunStoppedEarlyIsNamedWithItsSideAndItsOwnLength(t *testing.T) {
	t.Parallel()
	got := comparing(t, planned{seed: 4, endS: 120}, planned{seed: 4, endS: 120, stopS: 60})
	codes := limitCodes(got)
	if !slices.Contains(codes, "stopped_earlyb") || slices.Contains(codes, "stopped_earlya") || !slices.Contains(codes, "different_simulated_length") {
		t.Fatalf("limits = %v, want stopped_early for B and different_simulated_length", codes)
	}
	if got.A.SimulatedLengthNS != 120e9 || got.B.SimulatedLengthNS != 60e9 {
		t.Fatalf("lengths = %d and %d, want 120 s and 60 s", got.A.SimulatedLengthNS, got.B.SimulatedLengthNS)
	}
}

func TestADifferentDomainIsNamedAndItsPlanIsNotPairedWithAnotherDomainsFaults(t *testing.T) {
	t.Parallel()
	got := comparing(t, planned{seed: 4, fault: true, endS: 120}, planned{domain: "shipment-exception", seed: 4, endS: 120})
	if !slices.Contains(limitCodes(got), "different_domain") || len(got.Plan.Faults.OnlyInA) != 1 {
		t.Fatalf("limits = %v faults = %+v, want different_domain and the fault only in A", limitCodes(got), got.Plan.Faults)
	}
}

func TestAnOpenRunCannotBeComparedAndNamesTheSideAndThatItIsNotEnded(t *testing.T) {
	t.Parallel()
	open := plannedFolder(t, planned{seed: 4, endS: 120})
	if err := os.Remove(filepath.Join(open, "run.json")); err != nil {
		t.Fatal(err)
	}
	_, err := compare(t.Context(), plannedFolder(t, planned{seed: 4, endS: 120}), open)
	if !errors.Is(err, report.ErrRunNotEnded) || !strings.Contains(err.Error(), "run B") {
		t.Fatalf("err = %v, want ErrRunNotEnded naming run B", err)
	}
}

func TestAReplayedReportOfARealRunReproducesAndAStoppedRunAsksWhyItStopped(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "replays one run")
	dirs := report.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	stopped := plannedFolder(t, planned{seed: 4, fault: true, endS: 120, stopS: 60})
	built, err := buildVerified(t.Context(), stopped, dirs)
	if err != nil || built.Replay.Status != "reproduces" {
		t.Fatalf("replay = %+v, %v, want reproduces", built.Replay, err)
	}
	codes := []string{}
	for _, question := range built.OpenQuestions {
		codes = append(codes, question.Code)
	}
	if !slices.Contains(codes, "stopped_early") || slices.Contains(codes, "replay_not_checked") {
		t.Fatalf("questions = %v, want stopped_early and no replay question", codes)
	}
	text, err := report.Render(built, report.FormatMarkdown)
	if err != nil || !strings.Contains(string(text), "`stopped_early`") || !strings.Contains(string(text), "| Replay | reproduces |") {
		t.Fatalf("markdown = %s, %v, want the stop among the open questions and the replay line", text, err)
	}
}

func TestProduceBuildsTheReportOrTheComparisonItsRequestNames(t *testing.T) {
	t.Parallel()
	dir := plannedFolder(t, planned{seed: 4, endS: 60})
	for name, tc := range map[string]struct {
		request report.Request
		want    string
	}{
		"a report":     {report.Request{Dir: dir}, "# Run report: "},
		"a comparison": {report.Request{A: dir, B: dir}, "# Run comparison: "},
	} {
		document, err := report.Produce(t.Context(), tc.request)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		text, err := report.Render(document, report.FormatMarkdown)
		if err != nil || !strings.HasPrefix(string(text), tc.want) {
			t.Errorf("%s: markdown = %q (%v), want it to start with %q", name, text, err, tc.want)
		}
	}
	if _, err := report.Render(3, report.FormatJSON); !errors.Is(err, report.ErrUnknownDocument) {
		t.Errorf("rendering a number: %v, want ErrUnknownDocument", err)
	}
}
