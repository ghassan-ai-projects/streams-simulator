package domain

import (
	"strconv"
	"strings"
)

func ComparisonMarkdown(c *Comparison) []byte {
	var out strings.Builder
	out.WriteString("# Run comparison: " + titleID(c.A.RunID) + " and " + titleID(c.B.RunID) + "\n\nRun A is the reference. This document lists what differs; it does not rank the runs.\n")
	for _, s := range comparisonSections(c) {
		if len(s.lines) > 0 {
			out.WriteString("\n## " + s.title + "\n\n" + strings.Join(s.lines, "\n") + "\n")
		}
	}
	return []byte(out.String())
}

func comparisonSections(c *Comparison) []section {
	return []section{
		{"Inputs", inputLines(c)},
		{"Comparability limits", limitLines(c.Limits)},
		{"Plan differences", planDiffLines(c.Plan, c.A.StartTimeNS)},
		{"Results", resultLines(c)},
		{"Windowed results", []string{c.WindowedResults.Status + ": " + escapeText(c.WindowedResults.Reason)}},
	}
}

func inputLines(c *Comparison) []string {
	rows := [][]string{}
	for _, f := range inputFieldsOf(c.A, c.B) {
		if f.a != "" || f.b != "" {
			rows = append(rows, []string{f.name, f.a, f.b})
		}
	}
	return table([]string{"Field", "A", "B"}, rows)
}

type pairField struct{ name, a, b string }

func inputFieldsOf(a, b RunInputs) []pairField {
	return []pairField{
		{"Run ID", code(a.RunID), code(b.RunID)}, {"Name", a.Name, b.Name}, {"Purpose", a.Purpose, b.Purpose}, {"Mode", a.Mode, b.Mode},
		{"Seed", seedText(a.Seed), seedText(b.Seed)}, {"Simulator version", a.SimVersion, b.SimVersion},
		{"Domain", reference(a.Domain), reference(b.Domain)}, {"Adapter", reference(a.Adapter), reference(b.Adapter)},
		{"Time mode", a.TimeMode, b.TimeMode}, {"Start", instant(a.StartTimeNS, 0), instant(b.StartTimeNS, 0)},
		{"Requested end", instant(a.RequestedEndNS, a.StartTimeNS), instant(b.RequestedEndNS, b.StartTimeNS)},
		{"End", instant(a.EndTimeNS, a.StartTimeNS), instant(b.EndTimeNS, b.StartTimeNS)},
		{"Simulated length", clockText(a.SimulatedLengthNS), clockText(b.SimulatedLengthNS)},
		{"Termination", terminationText(Run{Termination: a.Termination, StartTimeNS: a.StartTimeNS}), terminationText(Run{Termination: b.Termination, StartTimeNS: b.StartTimeNS})},
		{"Pace", rateText(a.Rate), rateText(b.Rate)}, {"Complete", yesNo(a.Complete), yesNo(b.Complete)},
		{"Verification", a.Verification, b.Verification},
	}
}

func limitLines(limits []Limit) []string {
	lines := make([]string, len(limits))
	for i, limit := range limits {
		lines[i] = "- " + escapeCell(code(limit.Code)) + ": " + escapeText(limit.Message)
	}
	return lines
}

func planDiffLines(plan PlanDiff, start int64) []string {
	return planDiffSections(plan, start)[1:]
}

func planDiffSections(plan PlanDiff, start int64) []string {
	lines := []string{}
	lines = append(lines, diffSection("Faults", plan.Faults, []string{"Entity", "Fault", "Onset", "Cleared"}, func(f FaultItem) []string {
		return []string{f.EntityID, f.Fault, instant(f.OnsetNS, start), optionalInstant(f.ClearedNS, start)}
	})...)
	lines = append(lines, diffSection("Perturbations", plan.Perturbations, []string{"Action", "Perturbation", "At", "From", "Until", "Params"}, func(p PerturbationItem) []string {
		return []string{p.Action, first(p.Perturbation, p.PerturbID), instant(p.AtNS, start), optionalInstant(p.FromNS, start), optionalInstant(p.UntilNS, start), codeJSON(p.Params)}
	})...)
	lines = append(lines, diffSection("Environment", plan.Environment, []string{"At", "Target", "Fault", "Params"}, func(e EnvironmentItem) []string {
		return []string{instant(e.AtNS, start), e.Target, e.Fault, codeJSON(e.Params)}
	})...)
	return append(lines, diffSection("Effector calls", plan.EffectorCalls, []string{"At", "Entity", "Effector", "Args", "Outcome"}, func(c CallItem) []string {
		return []string{instant(c.AtNS, start), c.EntityID, c.Effector, codeJSON(c.Args), outcomeText(c.Outcome)}
	})...)
}

func diffSection[T any](title string, diff Diff[T], header []string, cells func(T) []string) []string {
	lines := []string{"", "### " + title, "", "Identical: " + strconv.Itoa(len(diff.Identical)) + ". Different: " + strconv.Itoa(len(diff.Different)) +
		". Only in A: " + strconv.Itoa(len(diff.OnlyInA)) + ". Only in B: " + strconv.Itoa(len(diff.OnlyInB)) + "."}
	lines = appendGroup(lines, "Only in A", header, diff.OnlyInA, cells)
	lines = appendGroup(lines, "Only in B", header, diff.OnlyInB, cells)
	return appendChanged(lines, header, diff.Different, cells)
}

func appendGroup[T any](lines []string, title string, header []string, items []T, cells func(T) []string) []string {
	if len(items) == 0 {
		return lines
	}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = cells(item)
	}
	return append(append(lines, "", "#### "+title, ""), table(header, rows)...)
}

func appendChanged[T any](lines []string, header []string, changed []Changed[T], cells func(T) []string) []string {
	if len(changed) == 0 {
		return lines
	}
	rows := [][]string{}
	for _, pair := range changed {
		rows = append(rows, append(append([]string{"A"}, cells(pair.A)...), strings.Join(pair.Differences, ", ")), append(append([]string{"B"}, cells(pair.B)...), ""))
	}
	return append(append(lines, "", "#### Different", ""), table(append(append([]string{"Run"}, header...), "Differs in"), rows)...)
}
