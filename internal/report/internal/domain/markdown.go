package domain

import "strings"

type section struct {
	title string
	lines []string
}

func Markdown(report *Report) []byte {
	var out strings.Builder
	out.WriteString("# Run report: " + titleID(report.Run.RunID) + "\n\n" + statusLine(report) + "\n")
	for _, s := range sectionsOf(report) {
		if len(s.lines) > 0 {
			out.WriteString("\n## " + s.title + "\n\n" + strings.Join(s.lines, "\n") + "\n")
		}
	}
	return []byte(out.String())
}

func titleID(id string) string {
	if id == "" {
		return "unknown"
	}
	return escapeText(id)
}

func statusLine(report *Report) string {
	run := report.Run
	return "mode: " + run.Mode + " | complete: " + yesNo(run.Complete) + " | ended: " + yesNo(run.Ended) +
		" | verification: " + report.Verification.Status
}

func sectionsOf(report *Report) []section {
	sections := []section{
		{"Purpose", purposeLines(report.Run)},
		{"What ran", runLines(report.Run)},
		{"Outcome and integrity", outcomeLines(report)},
		{"Evidence coverage", countLines(report.Counts)},
		{"Results", resultsLines(report)},
	}
	sections = append(sections, resultSections(report)...)
	sections = append(sections, section{"Plan", planLines(report)})
	sections = append(sections, planSections(report)...)
	return append(sections, closingSections(report)...)
}

func closingSections(report *Report) []section {
	return []section{
		{"Open questions", questionLines(report.OpenQuestions)},
		{"Digests", digestLines(report.Digests)},
		{"Verification", verificationLines(report.Verification)},
		{"Torn tail", tornLines(report.TornTail)},
		{"Consumer", consumerLines(report.Consumer)},
		{"Scoring", scoringLines(report.Scoring)},
	}
}
