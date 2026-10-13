package architecture

import (
	"strings"
	"testing"
)

const wiringBefore = `package architecture

const cliF = "internal/cli/internal/files"

var cliAppFiles = map[string]cliFile{
	"run.go":   {"command: run", []string{cliF, "internal/run"}, 95},
	"score.go": {"command: score", []string{"internal/grading"}, 70},
}

var cliEdges = map[string]cliEdge{
	cliF: {[]string{"internal/kernel"}, 40},
}

var cliStdlib = []string{"errors", "fmt"}
`

const dependenciesBefore = `package architecture

var packageDependencies = map[string]string{
	"internal/cli": "adapter run",
	"internal/run": "",
}
`

func ratchet(t *testing.T, before, after, table, list string) []string {
	t.Helper()
	tables, lists := []string{table}, []string{list}
	if table == "" {
		tables = nil
	}
	if list == "" {
		lists = nil
	}
	violations, err := goGateViolations("gate.go", []byte(before), []byte(after), tables, lists)
	if err != nil {
		t.Fatal(err)
	}
	return violations
}

func TestRatchetFlagsAGateThatGotLooser(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, from, to, table, list, want string
	}{
		{"a raised line cap", "95}", "96}", "cliAppFiles", "", "cap raised from 95 to 96"},
		{"a raised edge cap", "40}", "41}", "cliEdges", "", "cap raised from 40 to 41"},
		{"an added file import", `"internal/grading"}`, `"internal/grading", "internal/model"}`, "cliAppFiles", "", `"internal/model" added`},
		{"an added stdlib entry", `"errors", "fmt"`, `"errors", "fmt", "os"`, "", "cliStdlib", `"os" added`},
		{"a grown package allow-list", `"adapter run"`, `"adapter run kernel"`, "packageDependencies", "", `"kernel" added`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			before := wiringBefore + strings.TrimPrefix(dependenciesBefore, "package architecture")
			after := strings.Replace(before, tc.from, tc.to, 1)
			got := ratchet(t, before, after, tc.table, tc.list)
			if len(got) != 1 || !strings.Contains(got[0], tc.want) {
				t.Fatalf("violations = %v, want one containing %q", got, tc.want)
			}
		})
	}
}

func TestRatchetLetsAGateTightenAndNewKeysIn(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, from, to, table string }{
		{"a lowered cap", "95}", "80}", "cliAppFiles"},
		{"a removed import", `[]string{cliF, "internal/run"}`, `[]string{cliF}`, "cliAppFiles"},
		{"a new file with imports and a cap", `"score.go":`, `"extra.go": {"x", []string{"internal/model"}, 500},` + "\n\t" + `"score.go":`, "cliAppFiles"},
		{"a new dependency entry", `"internal/run": ""`, `"internal/run": "", "internal/new": "model"`, "packageDependencies"},
		{"a removed entry", `	"internal/run": "",` + "\n", "", "packageDependencies"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			before := wiringBefore + strings.TrimPrefix(dependenciesBefore, "package architecture")
			after := strings.Replace(before, tc.from, tc.to, 1)
			if got := ratchet(t, before, after, tc.table, ""); len(got) != 0 {
				t.Fatalf("violations = %v, want none", got)
			}
		})
	}
}

func TestRatchetRefusesATableThatWentMissing(t *testing.T) {
	t.Parallel()
	after := strings.Replace(wiringBefore, "cliAppFiles", "renamedFiles", 1)
	got := ratchet(t, wiringBefore, after, "cliAppFiles", "")
	if len(got) != 1 || !strings.Contains(got[0], "no longer found") {
		t.Fatalf("violations = %v, want the missing table reported", got)
	}
}

func TestRatchetTreatsAFileNewSinceTheBranchPointAsFree(t *testing.T) {
	t.Parallel()
	got, err := goGateViolations("gate.go", nil, []byte(wiringBefore), []string{"cliAppFiles"}, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("violations = %v, %v, want none for a file with no base version", got, err)
	}
}

func TestRatchetFlagsALoweredCoverageFloorButNotARaisedOrNewOne(t *testing.T) {
	t.Parallel()
	base := []byte("# comment\ninternal/a 90\ninternal/b 80\ncmd/streamsim 0\ninternal/gone 95\n")
	current := []byte("internal/a 85\ninternal/b 99\ninternal/new 10\ncmd/streamsim 0\n")
	exists := func(pkg string) bool { return pkg != "internal/gone" }
	got := coverageFloorViolations(base, current, exists)
	if len(got) != 1 || !strings.Contains(got[0], "internal/a: floor lowered from 90 to 85") {
		t.Fatalf("violations = %v, want only internal/a lowered", got)
	}
	dropped := coverageFloorViolations([]byte("internal/a 90\n"), []byte(""), func(string) bool { return true })
	if len(dropped) != 1 || !strings.Contains(dropped[0], "from 90 to 70") {
		t.Fatalf("violations = %v, want an unlisted package to fall back to the default floor", dropped)
	}
}

func TestRatchetFlagsAnAddedTestHookButNotARemovedOne(t *testing.T) {
	t.Parallel()
	base := []byte("# header\ninternal/a/operations.go:Load  reason one\ninternal/b/service.go:New  reason two\n")
	current := []byte("internal/a/operations.go:Load  reason reworded\ninternal/c/service.go:New  reason three\n")
	got := hookViolations(base, current)
	if len(got) != 1 || !strings.Contains(got[0], "internal/c/service.go:New: test hook added") {
		t.Fatalf("violations = %v, want only the added hook", got)
	}
}

func TestOnlyANamedReasonApprovesAGateChange(t *testing.T) {
	t.Parallel()
	messages := "feat: x\n\nOwner-approved-gate-change: the new module needs the director\n\nrefactor: y\nOwner-approved-gate-change:   \nOwner-approved-gate-change-not: nope\n"
	got := approvals(messages)
	if len(got) != 1 || got[0] != "the new module needs the director" {
		t.Fatalf("approvals = %v, want the one named reason", got)
	}
	if got := approvals("refactor: no approval\n\nCo-Authored-By: someone\n"); len(got) != 0 {
		t.Fatalf("approvals = %v, want none", got)
	}
}
