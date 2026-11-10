package domain

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var delimiterRow = regexp.MustCompile(`^\|(\s*:?-+:?\s*\|)+$`)

// tableProblems parses every Markdown table of a document: a block set off by blank lines whose first line is a header
// row, whose second is a delimiter row of the same width and whose other rows have that width and are no delimiter.
func tableProblems(markdown string) (tables int, problems []string) {
	for _, block := range strings.Split(markdown, "\n\n") {
		lines := strings.Split(strings.Trim(block, "\n"), "\n")
		if !strings.HasPrefix(lines[0], "|") {
			problems = append(problems, strayRows(lines)...)
			continue
		}
		tables++
		problems = append(problems, tableShape(lines)...)
	}
	return tables, problems
}

func strayRows(lines []string) []string {
	var problems []string
	for _, line := range lines {
		if strings.HasPrefix(line, "|") {
			problems = append(problems, "a table row not set off by a blank line: "+line)
		}
	}
	return problems
}

func tableShape(lines []string) []string {
	width := cellCount(lines[0])
	if len(lines) < 2 || !delimiterRow.MatchString(lines[1]) || cellCount(lines[1]) != width {
		return []string{"a table without its delimiter row: " + lines[0]}
	}
	var problems []string
	for _, line := range lines[2:] {
		switch {
		case delimiterRow.MatchString(line):
			problems = append(problems, "a second header inside the table of "+lines[0])
		case !strings.HasPrefix(line, "|") || cellCount(line) != width:
			problems = append(problems, "a row of another width under "+lines[0]+": "+line)
		}
	}
	return problems
}

func cellCount(row string) int {
	pipes, escaped := 0, false
	for _, r := range row {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == '|':
			pipes++
		}
	}
	return pipes - 1
}

func TestTheTableParserCatchesTwoTablesGluedTogether(t *testing.T) {
	t.Parallel()
	glued := "## Outcome score\n\n| Field | Value |\n|---|---|\n| Status | scored |\n| Field | Value |\n|---|---|\n| Valid | yes |\n"
	if _, problems := tableProblems(glued); len(problems) == 0 {
		t.Fatal("two tables without a blank line between them parse as one")
	}
	if _, problems := tableProblems("text\n| a | b |\n|---|---|\n"); len(problems) == 0 {
		t.Fatal("a table glued to a paragraph parses")
	}
	if tables, problems := tableProblems("| a \\| b | c |\n|---|---|\n| x | y |\n"); tables != 1 || len(problems) != 0 {
		t.Fatalf("an escaped pipe in a cell: %d tables, problems %q", tables, problems)
	}
}

func TestEveryTableOfAReportWithAnOutcomeScoreParses(t *testing.T) {
	t.Parallel()
	report := &Report{Run: Run{RunID: "r-1", Mode: "evaluate", Ended: true}, Verification: Verification{Status: FilesMatch},
		Truth: &Truth{KPIs: &model.KPITruth{Entities: []model.EntityKPIs{{EntityID: "B-1",
			KPIs: []model.KPIValue{{State: "delivered", Mean: 2, Min: 1, Max: 3, Samples: 3}}}}}},
		OutcomeScore: &OutcomeScoring{Status: Scored, Result: &model.OutcomeReport{OutcomeScore: model.OutcomeScore{Valid: true,
			Runs: []model.OutcomeRun{{Mode: model.OutcomeAsRun, Facts: &model.OutcomeFacts{}, Cost: &model.OutcomeCost{J: -1}},
				{Mode: model.OutcomeDoNothing, Facts: &model.OutcomeFacts{}, Cost: &model.OutcomeCost{J: 2}}},
			Verdicts: []model.OutcomeVerdict{{Baseline: model.OutcomeDoNothing, Delta: -3, Verdict: model.VerdictBeats}}}}}}
	tables, problems := tableProblems(string(Markdown(report)))
	for _, problem := range problems {
		t.Error(problem)
	}
	if tables < 6 {
		t.Fatalf("parsed %d tables, want the run, outcome, KPI, alarm system, plan and outcome score tables", tables)
	}
}
