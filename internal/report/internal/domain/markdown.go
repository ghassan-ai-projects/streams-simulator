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
	return id
}

func statusLine(report *Report) string {
	run := report.Run
	return "mode: " + run.Mode + " | complete: " + yesNo(run.Complete) + " | ended: " + yesNo(run.Ended) +
		" | verification: " + report.Verification.Status
}

func sectionsOf(report *Report) []section {
	sections := []section{
		{"Run", runLines(report.Run)},
		{"Digests", digestLines(report.Digests)},
		{"Verification", verificationLines(report.Verification)},
		{"Deliveries", countLines(report.Counts)},
	}
	sections = append(sections, truthSections(report)...)
	return append(sections, section{"Consumer", consumerLines(report.Consumer)}, section{"Scoring", scoringLines(report.Scoring)})
}
