package domain

import (
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var verdictOrder = []string{model.VerdictBeats, model.VerdictTies, model.VerdictLoses}

func outcomeScoreLines(scoring *OutcomeScoring) []string {
	if scoring == nil {
		return nil
	}
	if scoring.Result == nil {
		return fieldTable([]field{{"Status", scoring.Status}, {"Reason", scoring.Reason}})
	}
	lines := append([]string{outcomeSummary(scoring), "", costReading(scoring), ""}, fieldTable(outcomeFields(scoring))...)
	return append(lines, outcomeTables(scoring)...)
}

func outcomeTables(scoring *OutcomeScoring) []string {
	runHeaders := []string{"Run", "Policy", "Cost J" + unitSuffix(scoring.Unit), "Trips", "Boundary violations", "Calls"}
	lines := append([]string{""}, table(runHeaders, outcomeRunRows(scoring))...)
	if len(scoring.Result.Verdicts) == 0 {
		return lines
	}
	return append(append(lines, ""), table([]string{"Baseline", "Policy", "Delta J", "Verdict"}, verdictRows(scoring))...)
}

func outcomeSummary(scoring *OutcomeScoring) string {
	var clauses []string
	for _, verdict := range verdictOrder {
		if baselines := baselinesJudged(scoring.Result.Verdicts, verdict); len(baselines) > 0 {
			clauses = append(clauses, verdict+" "+listed(baselines))
		}
	}
	if len(clauses) == 0 {
		return "**Summary:** No verdict: the domain prices no outcome, so the run is compared with nothing."
	}
	return "**Summary:** The run " + strings.Join(clauses, "; it ") + "."
}

func baselinesJudged(verdicts []model.OutcomeVerdict, verdict string) []string {
	var baselines []string
	for _, judged := range verdicts {
		if judged.Verdict == verdict {
			baselines = append(baselines, escapeText(baselineName(judged.Baseline))+" ("+costDelta(judged.Delta)+")")
		}
	}
	return baselines
}

func baselineName(mode model.OutcomeMode) string {
	if mode == model.OutcomeDoNothing {
		return "doing nothing"
	}
	if name, ok := mode.Policy(); ok {
		return name
	}
	return string(mode)
}

func listed(items []string) string {
	if len(items) == 1 {
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func costReading(scoring *OutcomeScoring) string {
	unit := "the domain's own cost unit (the domain does not name it)"
	if scoring.Unit != "" {
		unit = escapeText(scoring.Unit)
	}
	reading := "Costs are J in " + unit + ". A delta is the run's J minus the baseline's: a negative delta means the run cost less."
	if tie := scoring.Result.Tie; tie != nil {
		reading += " Two costs within " + strconv.FormatFloat(*tie, 'f', -1, 64) + " of each other tie."
	}
	return reading
}

func outcomeFields(scoring *OutcomeScoring) []field {
	result := scoring.Result
	return []field{{"Status", scoring.Status}, {"Story", storyText(result.Valid)}, {"Easy", easyText(result.Valid, result.Easy)},
		{"Cost unit", scoring.Unit}}
}

func storyText(valid bool) string {
	if valid {
		return "Reaction story: doing nothing is not clean (a boundary is violated or a trip trips), so the story tests a reaction (" +
			code("valid: yes") + ")"
	}
	return "Control story: doing nothing is clean, so the story tests no reaction and an unneeded call only costs (" + code("valid: no") + ")"
}

func easyText(valid, easy bool) string {
	switch {
	case easy:
		return "yes: a baseline policy ties or beats the run, so the story does not tell the run from that fixed rule"
	case valid:
		return "no: the run beats every baseline policy"
	}
	return "no: a control story is never easy"
}

func unitSuffix(unit string) string {
	if unit == "" {
		return ""
	}
	return " (" + unit + ")"
}

func (scoring *OutcomeScoring) policyText(mode model.OutcomeMode) string {
	switch mode {
	case model.OutcomeAsRun:
		return "the run as it ran"
	case model.OutcomeDoNothing:
		return "nobody acts"
	}
	return scoring.Labels[string(mode)]
}
