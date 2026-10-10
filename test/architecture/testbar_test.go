package architecture

import (
	"go/ast"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestNoTestsAtRepositoryRoot(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			t.Errorf("%s: tests live with the module they prove or in test/", entry.Name())
		}
	}
}

// sleepingTests is the ratchet of tests that still call time.Sleep (T5). A
// file may only lower its count; the table is empty when the last one goes.
var sleepingTests = map[string]int{
	"internal/device/uds_test.go": 1,
}

func TestTestsNeverSleep(t *testing.T) {
	t.Parallel()
	counts := map[string]int{}
	for _, file := range testFiles(t) {
		ast.Inspect(file.source, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				if use, ok := selectorCapability(file, selector); ok && use == "call:time.Sleep" {
					counts[file.path]++
				}
			}
			return true
		})
	}
	for path, count := range counts {
		if count > sleepingTests[path] {
			t.Errorf("%s calls time.Sleep %d times (allowed %d): wait on a channel, condition or t.Context()", path, count, sleepingTests[path])
		}
	}
	for path, allowed := range sleepingTests {
		if counts[path] < allowed {
			t.Errorf("sleepingTests is stale: %s has %d sleeps, table says %d", path, counts[path], allowed)
		}
	}
}

// planningVocabulary matches test names that carry a round, phase or ticket
// number instead of describing what it proves (T3).
var planningVocabulary = regexp.MustCompile(`(Round|Phase|Wave|Stage|Level|Step|Milestone)\d|(^|[a-z])[SPL]\d+([A-Z_]|$)`)

func TestTestNamesCarryNoPlanningVocabulary(t *testing.T) {
	t.Parallel()
	for _, file := range testFiles(t) {
		for _, decl := range file.source.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !strings.HasPrefix(function.Name.Name, "Test") {
				continue
			}
			if planningVocabulary.MatchString(function.Name.Name) {
				t.Errorf("%s: test %s names a planning item; name what it proves", file.path, function.Name.Name)
			}
		}
	}
}

func TestPlanningVocabularyPatternIsSpecific(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]bool{
		"TestArtifactRoundTrip":              false,
		"TestReplayRoundTripsParams":         false,
		"TestRound3Rejects":                  true,
		"TestAdapterConformsS1Gate":          true,
		"TestLeaseExpiresAfterItsTTL":        false,
		"TestPhase2DeliversEverything":       true,
		"TestProfileP0Fails":                 true,
		"TestStepFunctionIntegratesLinearly": false,
	} {
		if got := planningVocabulary.MatchString(name); got != want {
			t.Errorf("%s: matches=%v, want %v", name, got, want)
		}
	}
}
