package architecture

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// The gates are only worth anything while they stay as strict as the day they
// were written. This ratchet compares the caps and allow-lists of every gate
// table with the branch point from main: a cap may fall, never rise; an
// allow-list may shrink, never grow; a coverage floor may rise, never fall; a
// test hook may be removed, never added. A key that did not exist at the
// branch point (a new package, a new file) is free. The owner can approve a
// loosening with a commit line `Owner-approved-gate-change: <reason>`
// anywhere between the branch point and HEAD.

const (
	approvalLine         = "Owner-approved-gate-change:"
	defaultCoverageFloor = 70
	ratchetRoot          = "../.."
)

var (
	ratchetedWiring = ratchetedGoFile{
		path:   "test/architecture/cli_wiring_test.go",
		tables: []string{"cliAppFiles", "cliEdges"},
		lists:  []string{"cliStdlib", "cliJSONFiles", "cliExported"},
	}
	ratchetedDependencies = ratchetedGoFile{
		path:   "test/architecture/dependencies_test.go",
		tables: []string{"packageDependencies"},
	}
)

type ratchetedGoFile struct {
	path          string
	tables, lists []string
}

func TestGateRatchetHoldsAgainstTheMainBranch(t *testing.T) {
	t.Parallel()
	base := ratchetBase(t)
	violations := gateViolations(t, base)
	if len(violations) == 0 {
		return
	}
	if reasons := approvals(commitMessages(t, base)); len(reasons) > 0 {
		t.Logf("gate loosening approved by the owner: %s", strings.Join(reasons, "; "))
		return
	}
	for _, violation := range violations {
		t.Error(violation)
	}
	t.Errorf("a gate got looser than at %s; move the code instead, or have the owner approve it with a commit line %q", base, approvalLine)
}

func ratchetBase(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("gate ratchet skipped: git is not available")
	}
	for _, main := range []string{"origin/main", "main"} {
		if base, err := git(t, "merge-base", main, "HEAD"); err == nil && base != "" {
			return base
		}
	}
	t.Skip("gate ratchet skipped: no origin/main or main branch to compare with (fetch the full history)")
	return ""
}

func git(t *testing.T, args ...string) (string, error) {
	t.Helper()
	command := exec.CommandContext(t.Context(), "git", append([]string{"-C", ratchetRoot}, args...)...) //nolint:gosec // G204: the arguments are constants and a commit id git printed
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}

func gateViolations(t *testing.T, base string) []string {
	t.Helper()
	var violations []string
	for _, gate := range []ratchetedGoFile{ratchetedWiring, ratchetedDependencies} {
		found, err := goGateViolations(gate.path, baseVersion(t, base, gate.path), currentVersion(t, gate.path), gate.tables, gate.lists)
		if err != nil {
			t.Fatal(err)
		}
		violations = append(violations, found...)
	}
	floors := coverageFloorViolations(baseVersion(t, base, "scripts/coverage-floors.txt"), currentVersion(t, "scripts/coverage-floors.txt"), packageExists)
	hooks := hookViolations(baseVersion(t, base, "scripts/test-hooks.txt"), currentVersion(t, "scripts/test-hooks.txt"))
	return append(append(violations, floors...), hooks...)
}

func baseVersion(t *testing.T, base, path string) []byte {
	t.Helper()
	content, err := git(t, "show", base+":"+path)
	if err != nil {
		return nil
	}
	return []byte(content + "\n")
}

func currentVersion(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(ratchetRoot + "/" + path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}

func packageExists(pkg string) bool {
	info, err := os.Stat(ratchetRoot + "/" + pkg)
	return err == nil && info.IsDir()
}

func commitMessages(t *testing.T, base string) string {
	t.Helper()
	messages, err := git(t, "log", "--format=%B", base+"..HEAD")
	if err != nil {
		t.Fatalf("read the commit messages since %s: %v", base, err)
	}
	return messages
}

func approvals(messages string) []string {
	var reasons []string
	for _, line := range strings.Split(messages, "\n") {
		if reason, found := strings.CutPrefix(strings.TrimSpace(line), approvalLine); found && strings.TrimSpace(reason) != "" {
			reasons = append(reasons, strings.TrimSpace(reason))
		}
	}
	return reasons
}

func coverageFloorViolations(base, current []byte, exists func(string) bool) []string {
	before, now := floorsOf(base), floorsOf(current)
	var violations []string
	for _, pkg := range sortedKeys(before) {
		floor, listed := now[pkg]
		if !listed {
			floor = defaultCoverageFloor
		}
		if floor < before[pkg] && (listed || exists(pkg)) {
			violations = append(violations, fmt.Sprintf("scripts/coverage-floors.txt %s: floor lowered from %d to %d", pkg, before[pkg], floor))
		}
	}
	return violations
}

func floorsOf(content []byte) map[string]int {
	floors := map[string]int{}
	for _, fields := range dataLines(content) {
		if floor, err := strconv.Atoi(fields[len(fields)-1]); err == nil && len(fields) == 2 {
			floors[fields[0]] = floor
		}
	}
	return floors
}

func hookViolations(base, current []byte) []string {
	known := map[string]bool{}
	for _, fields := range dataLines(base) {
		known[fields[0]] = true
	}
	var violations []string
	for _, fields := range dataLines(current) {
		if !known[fields[0]] {
			violations = append(violations, fmt.Sprintf("scripts/test-hooks.txt %s: test hook added", fields[0]))
		}
	}
	return violations
}

func dataLines(content []byte) [][]string {
	var lines [][]string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		if fields := strings.Fields(scanner.Text()); len(fields) > 0 && !strings.HasPrefix(fields[0], "#") {
			lines = append(lines, fields)
		}
	}
	return lines
}
