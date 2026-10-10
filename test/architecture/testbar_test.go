package architecture

import (
	"go/ast"
	"go/token"
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

func TestTestsNeverSleep(t *testing.T) {
	t.Parallel()
	for _, file := range testFiles(t) {
		ast.Inspect(file.source, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				if use, ok := selectorCapability(file, selector); ok && use == "call:time.Sleep" {
					t.Errorf("%s: time.Sleep in a test: wait on a channel, condition or t.Context()", file.path)
				}
			}
			return true
		})
	}
}

func TestEveryTestRunsInParallel(t *testing.T) {
	t.Parallel()
	for _, file := range testFiles(t) {
		for _, decl := range file.source.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !isTopLevelTest(function) || optsOutOfParallel(function) {
				continue
			}
			if !startsInParallel(function) {
				t.Errorf("%s: %s must call t.Parallel() first (or say why not with //nolint:paralleltest // <reason>)", file.path, function.Name.Name)
			}
		}
	}
}

func isTopLevelTest(function *ast.FuncDecl) bool {
	params := function.Type.Params.List
	if !strings.HasPrefix(function.Name.Name, "Test") || function.Name.Name == "TestMain" || len(params) != 1 {
		return false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "T"
}

func optsOutOfParallel(function *ast.FuncDecl) bool {
	if function.Doc == nil {
		return false
	}
	for _, comment := range function.Doc.List {
		if strings.Contains(comment.Text, "nolint:paralleltest") {
			return true
		}
	}
	return false
}

func startsInParallel(function *ast.FuncDecl) bool {
	if function.Body == nil || len(function.Body.List) == 0 {
		return false
	}
	statement, ok := function.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := statement.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Parallel"
}

// fixturePathLiteral matches a string literal that names a shipped domain,
// adapter or documentation example by path (T9: such a path is defined once,
// in internal/testsupport).
var fixturePathLiteral = regexp.MustCompile(`(^|/)(domains|adapters)/|/(domains|adapters)$|docs/examples/`)

func TestFixturePathsAreDefinedOnce(t *testing.T) {
	t.Parallel()
	for _, file := range testFiles(t) {
		if strings.HasPrefix(file.path, "internal/testsupport/") || strings.HasPrefix(file.path, "test/architecture/") {
			continue
		}
		ast.Inspect(file.source, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.BasicLit:
				if n.Kind == token.STRING && !strings.Contains(n.Value, "://") && fixturePathLiteral.MatchString(strings.Trim(n.Value, "\"`")) {
					t.Errorf("%s: %s names a shipped fixture by path; use internal/testsupport", file.path, n.Value)
				}
			case *ast.CallExpr:
				if joinsFixtureDirectory(file, n) {
					t.Errorf("%s: filepath.Join names a shipped fixture directory; use internal/testsupport", file.path)
				}
			}
			return true
		})
	}
}

// joinsFixtureDirectory reports a filepath.Join call with a "domains" or
// "adapters" path element.
func joinsFixtureDirectory(file productionFile, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Join" {
		return false
	}
	if ident, ok := selector.X.(*ast.Ident); !ok || file.imports[ident.Name] != "path/filepath" {
		return false
	}
	for _, arg := range call.Args {
		if lit, ok := arg.(*ast.BasicLit); ok && (lit.Value == `"domains"` || lit.Value == `"adapters"`) {
			return true
		}
	}
	return false
}

func TestProductionNeverImportsTestSupport(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		if file.pkgDir == "internal/testsupport" {
			continue
		}
		for _, imported := range modulePackages(file) {
			if imported == "internal/testsupport" {
				t.Errorf("%s imports internal/testsupport, which is for tests only", file.path)
			}
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

// unnamedErrorAssertions is the ratchet of tests that only assert an error
// occurred (`if err == nil { t.Fatal }`) without saying which (T4). A file may
// only lower its count; the table is empty when the last one is fixed.
var unnamedErrorAssertions = map[string]int{
	"internal/canonical/canonical_test.go":                       4,
	"internal/device/device_test.go":                             3,
	"internal/device/internal/domain/capabilities_test.go":       1,
	"internal/device/internal/domain/codec_test.go":              1,
	"internal/device/internal/domain/conformance_test.go":        1,
	"internal/device/internal/domain/fault_schedule_test.go":     1,
	"internal/device/internal/uds/uds_test.go":                   1,
	"internal/device/uds_test.go":                                1,
	"internal/deviceworld/deviceworld_test.go":                   1,
	"internal/deviceworld/internal/domain/bindings_test.go":      3,
	"internal/domain/internal/domain/domain_test.go":             2,
	"internal/mcp/internal/app/capability_test.go":               3,
	"internal/model/model_test.go":                               2,
	"internal/perturb/internal/domain/perturb_test.go":           1,
	"internal/perturb/perturb_test.go":                           2,
	"internal/refconsumer/internal/domain/process_steps_test.go": 1,
	"internal/run/internal/app/artifact_test.go":                 2,
	"internal/run/internal/app/sink_equivalence_test.go":         2,
	"internal/run/internal/domain/rules_test.go":                 1,
	"internal/run/run_test.go":                                   2,
	"internal/sink/internal/files/file_test.go":                  1,
	"internal/sink/internal/httppush/httppush_test.go":           1,
	"internal/truth/internal/domain/truth_test.go":               5,
	"internal/truth/truth_test.go":                               2,
	"internal/world/internal/domain/effector_order_test.go":      1,
	"internal/world/internal/domain/effector_test.go":            4,
	"internal/world/world_test.go":                               2,
}

func TestErrorAssertionsNameTheErrorTheyExpect(t *testing.T) {
	t.Parallel()
	counts := map[string]int{}
	for _, file := range testFiles(t) {
		if strings.HasPrefix(file.path, "test/architecture/") {
			continue
		}
		ast.Inspect(file.source, func(node ast.Node) bool {
			if statement, ok := node.(*ast.IfStmt); ok && assertsOnlyThatAnErrorOccurred(statement) {
				counts[file.path]++
			}
			return true
		})
	}
	for path, count := range counts {
		if count > unnamedErrorAssertions[path] {
			t.Errorf("%s: %d tests assert only that an error occurred (allowed %d): check errors.Is/As or the message", path, count, unnamedErrorAssertions[path])
		}
	}
	for path, allowed := range unnamedErrorAssertions {
		if counts[path] < allowed {
			t.Errorf("unnamedErrorAssertions is stale: %s has %d, table says %d", path, counts[path], allowed)
		}
	}
}

// assertsOnlyThatAnErrorOccurred matches `if err == nil { t.Fatal...(...) }`:
// the test expected a failure and does not look at which one.
func assertsOnlyThatAnErrorOccurred(statement *ast.IfStmt) bool {
	cond, ok := statement.Cond.(*ast.BinaryExpr)
	if !ok || cond.Op != token.EQL || statement.Else != nil || len(statement.Body.List) == 0 {
		return false
	}
	name, isIdent := cond.X.(*ast.Ident)
	nilLit, isNil := cond.Y.(*ast.Ident)
	if !isIdent || !isNil || nilLit.Name != "nil" || !strings.HasSuffix(strings.ToLower(name.Name), "err") {
		return false
	}
	call, ok := statement.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	invocation, ok := call.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := invocation.Fun.(*ast.SelectorExpr)
	return ok && (selector.Sel.Name == "Fatal" || selector.Sel.Name == "Fatalf" || selector.Sel.Name == "Error" || selector.Sel.Name == "Errorf")
}
