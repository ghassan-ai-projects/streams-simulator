package domain

import (
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func endedReport() *Report {
	observed, emitted := int64(10), int64(10)
	return &Report{
		SchemaVersion: SchemaVersion,
		Run:           Run{RunID: "r-1", Mode: "explore", Complete: true, Ended: true, StartTimeNS: 1_000_000_000_000_000_000},
		Verification:  Verification{Status: FilesMatch, Checks: []Check{}},
		Replay:        Replay{Status: ReplayReproduces},
		Counts:        Counts{Emitted: &emitted, Observed: &observed},
		Truth:         &Truth{KPIs: &model.KPITruth{}},
		Scoring:       &Scoring{Status: Scored},
	}
}

func codesOf(questions []Question) []string {
	codes := []string{}
	for _, question := range questions {
		codes = append(codes, question.Code)
	}
	return codes
}

func TestOpenQuestionsAreAskedByRulesOverTheReport(t *testing.T) {
	t.Parallel()
	zero := int64(0)
	cases := []struct {
		name   string
		change func(*Report)
		want   []string
	}{
		{"a clean scored run asks nothing", func(*Report) {}, []string{}},
		{"an incomplete run", func(r *Report) { r.Run.Complete, r.Run.IncompleteReason = false, "the sink failed" }, []string{"run_incomplete"}},
		{"a run stopped early", func(r *Report) {
			r.Run.RequestedEndNS = r.Run.StartTimeNS + 1800e9
			r.Run.Termination = &model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: r.Run.StartTimeNS + 540e9}
		}, []string{"stopped_early"}},
		{"files that differ from the recorded digests", func(r *Report) {
			r.Verification = Verification{Status: FilesDiffer, Checks: []Check{{Name: "trace", Status: CheckDiffers}}}
		}, []string{"files_differ"}},
		{"a replay nobody ran", func(r *Report) { r.Replay = Replay{Status: NotChecked} }, []string{"replay_not_checked"}},
		{"a replay that does not reproduce", func(r *Report) {
			r.Replay = Replay{Status: ReplayDiverges, Detail: "replayed 3 records"}
		}, []string{"replay_does_not_reproduce"}},
		{"no observations", func(r *Report) { r.Counts.Observed = &zero }, []string{"no_observations"}},
		{"no observer journal", func(r *Report) { r.Counts.Observed = nil }, []string{"no_observations"}},
		{"a torn tail", func(r *Report) { r.TornTail = []string{"ledger.jsonl"} }, []string{"torn_tail"}},
		{"no KPI results", func(r *Report) { r.Truth.KPIs = nil }, []string{"no_kpi_results"}},
		{"no outside scoring", func(r *Report) { r.Scoring = nil }, []string{"no_outside_scoring"}},
		{"an unsealed evaluate run", func(r *Report) {
			r.Run.Mode, r.Run.Ended, r.Run.Complete, r.Truth, r.TruthUnavailable = "evaluate", false, false, nil, TruthSealed
			r.Replay, r.Scoring = Replay{Status: NotChecked}, nil
		}, []string{"evaluate_unsealed", "no_outside_scoring"}},
		{"an explore run that is still open", func(r *Report) { r.Run.Ended, r.Run.Complete = false, false }, []string{"run_not_ended"}},
		{"a file of another run", func(r *Report) {
			r.Run.Ended, r.Run.Complete, r.Truth, r.TruthUnavailable = false, false, nil, foreignArtifact+"run_id differ from journal.jsonl"
			r.Replay = Replay{Status: NotChecked}
		}, []string{"evidence_foreign"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			report := endedReport()
			tc.change(report)
			got := codesOf(questionsOf(report))
			for _, want := range tc.want {
				if !slices.Contains(got, want) {
					t.Fatalf("questions = %v, want %q among them", got, want)
				}
			}
			if len(tc.want) == 0 && len(got) != 0 {
				t.Fatalf("questions = %v, want none", got)
			}
		})
	}
}

func TestAnOpenRunDoesNotAskForAReplayItCannotHave(t *testing.T) {
	t.Parallel()
	report := endedReport()
	report.Run.Ended, report.Run.Complete, report.Replay = false, false, Replay{Status: NotChecked}
	if got := codesOf(questionsOf(report)); slices.Contains(got, "replay_not_checked") {
		t.Fatalf("questions = %v, want no replay question for a run that has not ended", got)
	}
}

func TestOpenQuestionsComeInTheOrderOfTheRulesWithStableSentences(t *testing.T) {
	t.Parallel()
	report := endedReport()
	report.Run.Complete, report.Run.IncompleteReason = false, "the sink failed"
	report.Run.RequestedEndNS = report.Run.StartTimeNS + 1800e9
	report.Run.Termination = &model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: report.Run.StartTimeNS + 540e9}
	report.Replay, report.Scoring = Replay{Status: NotChecked}, nil
	got := questionsOf(report)
	want := []Question{
		{"run_incomplete", "The run ended incomplete: the sink failed. Its evidence is partial."},
		{"stopped_early", "The operator stopped the run at +540s of the requested +1800s; actions planned after the stop never ran."},
		{"replay_not_checked", "The recording was not replayed: file digests agree with the run, but nobody has shown that the run reproduces."},
		{"no_outside_scoring", "This is an unscored investigation: no outside system was scored against the run."},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("questions = %+v, want %+v", got, want)
	}
}

func TestMarkdownLeadsWithPurposeAndKeepsOutcomeAndIntegrityOnSeparateLines(t *testing.T) {
	t.Parallel()
	report := endedReport()
	report.Run.Name, report.Run.Purpose = "Cooling loss", "Does the setpoint change help?"
	report.Run.Termination = &model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: report.Run.StartTimeNS + 540e9}
	text := string(Markdown(WithReplay(report, Replay{Status: NotChecked})))
	order := []string{"## Purpose", "## What ran", "## Outcome and integrity", "## Evidence coverage", "## Results", "## Plan", "## Open questions"}
	last := -1
	for _, heading := range order {
		at := strings.Index(text, heading+"\n")
		if at < 0 || at < last {
			t.Fatalf("heading %q at %d after %d in:\n%s", heading, at, last, text)
		}
		last = at
	}
	for _, line := range []string{"| Execution outcome | Stopped by operator at ", "| Recording integrity | files\\_match |", "| Replay | not\\_checked |",
		"- `replay_not_checked`: ", "- `stopped_early`: ", "| Objective | Not assessed."} {
		if !strings.Contains(text, line) {
			t.Errorf("markdown lacks %q:\n%s", line, text)
		}
	}
}
