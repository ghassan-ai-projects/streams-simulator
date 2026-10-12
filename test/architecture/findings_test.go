package architecture

import (
	"strings"
	"testing"
)

// The simulation-canvas goal tracks every review finding in one register and
// its phases in one plan (docs/simulation-canvas/QUALITY_BAR.md §4). These
// gates keep the register honest: a finding assigned to finished work cannot
// stay open, and every assignment names real work.
const (
	findingsRegister = "docs/simulation-canvas/FINDINGS.md"
	findingsPlan     = "docs/simulation-canvas/PLAN.md"
	nextGoal         = "next-goal"
)

var findingStatuses = map[string]bool{
	"open": true, "done": true, nextGoal: true, "dropped": true, "not-a-defect": true,
}

type finding struct{ id, phase, status string }

func TestNoFindingOfFinishedWorkIsOpen(t *testing.T) {
	t.Parallel()
	for _, f := range openFindingsOfFinishedWork(registeredFindings(t), finishedPlanEntries(t)) {
		t.Errorf("%s: finding %s is open but %s is marked Done in %s", findingsRegister, f.id, f.phase, findingsPlan)
	}
}

func TestOpenFindingOfFinishedWorkIsReported(t *testing.T) {
	t.Parallel()
	findings := []finding{{"A", "P1", "open"}, {"B", "P1", "done"}, {"C", "P2", "open"}, {"D", nextGoal, nextGoal}}
	got := openFindingsOfFinishedWork(findings, map[string]bool{"P1": true})
	if len(got) != 1 || got[0].id != "A" {
		t.Fatalf("openFindingsOfFinishedWork = %v, want only A", got)
	}
}

func TestEveryFindingNamesRealWorkAndAKnownStatus(t *testing.T) {
	t.Parallel()
	known := plannedEntries(t)
	seen := map[string]bool{}
	for _, f := range registeredFindings(t) {
		if seen[f.id] {
			t.Errorf("%s: finding %s is listed twice", findingsRegister, f.id)
		}
		seen[f.id] = true
		if !findingStatuses[f.status] {
			t.Errorf("%s: finding %s has unknown status %q", findingsRegister, f.id, f.status)
		}
		if f.phase != nextGoal && f.phase != "—" && !known[f.phase] {
			t.Errorf("%s: finding %s names %q, which is not a row of %s", findingsRegister, f.id, f.phase, findingsPlan)
		}
		if (f.phase == nextGoal) != (f.status == nextGoal) {
			t.Errorf("%s: finding %s: phase %q and status %q disagree about the next goal", findingsRegister, f.id, f.phase, f.status)
		}
	}
	if len(seen) == 0 {
		t.Fatalf("%s lists no findings", findingsRegister)
	}
}

func TestFindingsRegisterParsesTheDocumentedShape(t *testing.T) {
	t.Parallel()
	rows := tableRows("| ID | Source | Sev | Summary | Phase | Status |\n|---|---|---|---|---|---|\n| X1 | R | Low | a \\| b | P3 | open |\n")
	want := []string{"X1", "R", "Low", "a | b", "P3", "open"}
	if len(rows) != 2 || strings.Join(rows[1], ",") != strings.Join(want, ",") {
		t.Fatalf("tableRows parsed %q, want the header and the row %q", rows, want)
	}
}

// openFindingsOfFinishedWork returns the open findings assigned to a finished entry.
func openFindingsOfFinishedWork(findings []finding, finished map[string]bool) []finding {
	var out []finding
	for _, f := range findings {
		if f.status == "open" && finished[f.phase] {
			out = append(out, f)
		}
	}
	return out
}

// registeredFindings reads every six-column row of the findings register.
func registeredFindings(t *testing.T) []finding {
	t.Helper()
	var out []finding
	for _, cells := range tableRows(readRepositoryText(t, findingsRegister)) {
		if len(cells) == 6 && cells[0] != "ID" {
			out = append(out, finding{id: cells[0], phase: cells[4], status: cells[5]})
		}
	}
	return out
}

// plannedEntries is the set of ids in the first column of the plan's phase table.
func plannedEntries(t *testing.T) map[string]bool {
	t.Helper()
	known := map[string]bool{}
	for _, cells := range tableRows(readRepositoryText(t, findingsPlan)) {
		if len(cells) == 5 && strings.HasPrefix(cells[0], "P") && cells[0] != "Phase" {
			known[cells[0]] = true
		}
	}
	if len(known) == 0 {
		t.Fatalf("%s has no phase table", findingsPlan)
	}
	return known
}

// finishedPlanEntries is the set of ids the progress log marks Done.
func finishedPlanEntries(t *testing.T) map[string]bool {
	t.Helper()
	done := map[string]bool{}
	for _, cells := range tableRows(readRepositoryText(t, findingsPlan)) {
		if len(cells) == 5 && cells[1] == "Done" {
			done[cells[0]] = true
		}
	}
	return done
}

// tableRows splits the Markdown table rows of text into trimmed cells,
// skipping separator rows. An escaped pipe stays inside its cell.
func tableRows(text string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.HasPrefix(line, "|---") {
			continue
		}
		rows = append(rows, splitCells(line))
	}
	return rows
}

func splitCells(line string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
	parts := strings.Split(strings.ReplaceAll(inner, `\|`, "\x00"), "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(strings.ReplaceAll(part, "\x00", "|"))
	}
	return cells
}

func readRepositoryText(t *testing.T, path string) string {
	t.Helper()
	raw, err := repositoryRoot(t).ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
