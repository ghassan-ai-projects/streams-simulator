package domain

import (
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func purposeLines(run Run) []string {
	lines := fieldTable([]field{{"Name", run.Name}, {"Purpose", run.Purpose}})
	if len(lines) == 0 {
		return []string{"No name or purpose was recorded for this run."}
	}
	return lines
}

func outcomeLines(report *Report) []string {
	return fieldTable([]field{
		{"Execution outcome", executionOutcome(report.Run)},
		{"Recording integrity", integrityText(report.Verification)},
		{"Replay", replayText(report.Replay)},
		{"Outside scoring", scoringText(report.Scoring)},
	})
}

func executionOutcome(run Run) string {
	switch {
	case !run.Ended:
		return "Not ended: no run.json"
	case !run.Complete:
		return "Ended incomplete: " + run.IncompleteReason
	case run.Termination == nil:
		return "Ended at " + instant(run.EndTimeNS, run.StartTimeNS)
	}
	return terminationText(run) + requestedText(run)
}

func requestedText(run Run) string {
	if run.RequestedEndNS == 0 {
		return ""
	}
	return ", requested end " + instant(run.RequestedEndNS, run.StartTimeNS)
}

func integrityText(verification Verification) string {
	text := verification.Status
	for _, check := range verification.Checks {
		if check.Status != CheckMatch {
			text += " (" + check.Name + ": " + check.Status + ")"
		}
	}
	return text
}

func replayText(replay Replay) string {
	return strings.TrimSpace(replay.Status + " " + replay.Detail)
}

func scoringText(scoring *Scoring) string {
	if scoring == nil {
		return "none"
	}
	return strings.TrimSpace(scoring.Status + " " + scoring.Reason)
}

func resultsLines(report *Report) []string {
	if report.Truth == nil {
		return nil
	}
	lines := []string{}
	if kpis := report.Truth.KPIs; kpis != nil {
		lines = append(lines, "Scope: whole run, "+offset(kpis.ToNS, kpis.FromNS)[1:]+" of simulated time.", "")
	}
	return append(lines, fieldTable([]field{{"Objective", "Not assessed. The run declares no objective, so nothing is judged against one."}})...)
}

func planLines(report *Report) []string {
	truth := report.Truth
	if truth == nil {
		return nil
	}
	return fieldTable([]field{
		{"Faults", countOf(len(truth.Faults), "cleared", clearedFaults(truth.Faults))},
		{"Perturbations", strconv.Itoa(len(truth.Perturbations))},
		{"Environment changes", strconv.Itoa(len(truth.Environment))},
		{"Effector calls", countOf(len(truth.EffectorCalls), "refused", refusedCalls(truth.EffectorCalls))},
	})
}

func countOf(total int, what string, some int) string {
	return strconv.Itoa(total) + " (" + strconv.Itoa(some) + " " + what + ")"
}

func clearedFaults(faults []model.FaultInterval) int {
	cleared := 0
	for _, fault := range faults {
		if fault.ClearedNS != nil {
			cleared++
		}
	}
	return cleared
}

func refusedCalls(calls []EffectorCall) int {
	refused := 0
	for _, call := range calls {
		if call.Outcome != nil && call.Outcome.Refusal != "" {
			refused++
		}
	}
	return refused
}

func questionLines(questions []Question) []string {
	if len(questions) == 0 {
		return []string{"None."}
	}
	lines := make([]string, len(questions))
	for i, question := range questions {
		lines[i] = "- " + escapeCell(code(question.Code)) + ": " + escapeText(question.Message)
	}
	return lines
}
