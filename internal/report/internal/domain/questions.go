package domain

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Question struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type rule func(*Report) (Question, bool)

var questionRules = []rule{evidenceForeign, evaluateUnsealed, runNotEnded, runIncomplete, stoppedEarly, filesDiffer,
	replayNotChecked, replayDiverges, noObservations, tornTail, noKPIResults, noOutsideScoring}

func questionsOf(report *Report) []Question {
	questions := []Question{}
	for _, ask := range questionRules {
		if question, asked := ask(report); asked {
			questions = append(questions, question)
		}
	}
	return questions
}

func foreignEvidence(report *Report) bool {
	return report.TruthUnavailable != "" && report.TruthUnavailable != TruthSealed
}

func evidenceForeign(report *Report) (Question, bool) {
	return Question{"evidence_foreign", "A file of the folder belongs to another run, so no truth is shown: " + report.TruthUnavailable},
		foreignEvidence(report)
}

func evaluateUnsealed(report *Report) (Question, bool) {
	open := !report.Run.Ended && !foreignEvidence(report) && report.Run.Mode == journal.ModeEvaluate
	return Question{"evaluate_unsealed", "The run has not ended, so its truth is still sealed: this report withholds the faults, calls and results and cannot be compared with another run."}, open
}

func runNotEnded(report *Report) (Question, bool) {
	open := !report.Run.Ended && !foreignEvidence(report) && report.Run.Mode == journal.ModeExplore
	return Question{"run_not_ended", "The run has not ended: its record is still open and its results are not final."}, open
}

func runIncomplete(report *Report) (Question, bool) {
	run := report.Run
	return Question{"run_incomplete", "The run ended incomplete: " + run.IncompleteReason + ". Its evidence is partial."}, run.Ended && !run.Complete
}

func stoppedEarly(report *Report) (Question, bool) {
	run := report.Run
	if run.Termination == nil || run.Termination.Reason != model.TerminationStoppedByOperator {
		return Question{}, false
	}
	return Question{"stopped_early", "The operator stopped the run at " + clockText(run.Termination.AtNS-run.StartTimeNS) + requested(run) +
		"; actions planned after the stop never ran."}, true
}

func requested(run Run) string {
	if run.RequestedEndNS == 0 {
		return ""
	}
	return " of the requested " + clockText(run.RequestedEndNS-run.StartTimeNS)
}

func filesDiffer(report *Report) (Question, bool) {
	var names []string
	for _, check := range report.Verification.Checks {
		if check.Status != CheckMatch {
			names = append(names, check.Name)
		}
	}
	return Question{"files_differ", "The files of the folder no longer agree with the digests the run recorded (" + strings.Join(names, ", ") + ")."},
		report.Verification.Status == FilesDiffer
}

func replayNotChecked(report *Report) (Question, bool) {
	message := "The recording was not replayed: file digests agree with the run, but nobody has shown that the run reproduces."
	if report.Replay.Detail != "" {
		message += " " + report.Replay.Detail
	}
	return Question{"replay_not_checked", message}, report.Run.Ended && report.Replay.Status == NotChecked
}

func replayDiverges(report *Report) (Question, bool) {
	return Question{"replay_does_not_reproduce", "Replaying the recording does not reproduce the run: " + report.Replay.Detail},
		report.Replay.Status == ReplayDiverges
}

func noObservations(report *Report) (Question, bool) {
	observed := report.Counts.Observed
	return Question{"no_observations", "The consumer's journal holds no observations: nothing a consumer could have seen is recorded."},
		observed == nil || *observed == 0
}

func tornTail(report *Report) (Question, bool) {
	return Question{"torn_tail", "The last line of " + strings.Join(report.TornTail, ", ") + " was cut short, as by a crash, and was dropped."},
		len(report.TornTail) > 0
}

func noKPIResults(report *Report) (Question, bool) {
	return Question{"no_kpi_results", "The run records no whole-run KPI results."}, report.Run.Ended && report.Truth != nil && report.Truth.KPIs == nil
}

func noOutsideScoring(report *Report) (Question, bool) {
	return Question{"no_outside_scoring", scoringSentence(report)}, report.Scoring == nil || report.Scoring.Status != Scored
}

func scoringSentence(report *Report) string {
	switch {
	case report.Scoring != nil && report.Scoring.Reason != "":
		return "The run was not scored: " + report.Scoring.Reason
	case report.Consumer != nil:
		return "A verdict is recorded but no label was available to score it."
	case report.Run.Mode == journal.ModeExplore:
		return "This is an unscored investigation: no outside system was scored against the run."
	}
	return "No outside system scored this run: no verdict was submitted."
}
