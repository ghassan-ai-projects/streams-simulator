package architecture

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strings"
	"testing"
)

const functionReviewLines = 60

// These cohesive operations were reviewed in REVIEW.md. Growth requires another
// review; an exception does not grant an unlimited function-size allowance.
var reviewedFunctions = map[string]int{
	"internal/canonical/encoding.go:writeValue":   66,
	"internal/mcp/schemas.go:toolSchema":          185,
	"internal/world/integration.go:World.rk4Step": 72,
}

func TestProductionFunctionsHaveReviewDecisions(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	seen := map[string]bool{}
	err = fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect functions in %s: %w", path, err)
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return fs.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		source, err := root.ReadFile(path)
		if err != nil {
			return fmt.Errorf("inspect functions in %s: %w", path, err)
		}
		positions := token.NewFileSet()
		file, err := parser.ParseFile(positions, path, source, 0)
		if err != nil {
			return fmt.Errorf("inspect functions in %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			key := path + ":" + functionName(fn)
			seen[key] = true
			lines := positions.Position(fn.Body.End()).Line - positions.Position(fn.Body.Pos()).Line + 1
			if !functionLengthReviewed(key, lines) {
				t.Errorf("%s has a %d-line body: simplify it or record a review decision", key, lines)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncLit)
			if !ok {
				return true
			}
			start := positions.Position(fn.Body.Pos())
			lines := positions.Position(fn.Body.End()).Line - start.Line + 1
			if lines > functionReviewLines {
				t.Errorf("%s:%d has a %d-line anonymous function: extract a named operation", path, start.Line, lines)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for key := range reviewedFunctions {
		if !seen[key] {
			t.Errorf("stale function review: %s", key)
		}
	}
}

func functionName(fn *ast.FuncDecl) string {
	if fn.Recv == nil {
		return fn.Name.Name
	}
	receiver := fn.Recv.List[0].Type
	if pointer, ok := receiver.(*ast.StarExpr); ok {
		receiver = pointer.X
	}
	if name, ok := receiver.(*ast.Ident); ok {
		return name.Name + "." + fn.Name.Name
	}
	return fmt.Sprintf("%T.%s", receiver, fn.Name.Name)
}

func functionLengthReviewed(key string, lines int) bool {
	return lines <= functionReviewLines || lines <= reviewedFunctions[key]
}

func TestFunctionReviewRejectsUnreviewedGrowth(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		lines    int
		approved bool
	}{
		{"internal/world/new.go:operation", 60, true},
		{"internal/world/new.go:operation", 61, false},
		{"internal/world/integration.go:World.rk4Step", 72, true},
		{"internal/world/integration.go:World.rk4Step", 73, false},
		{"internal/world/integration.go:Other.rk4Step", 72, false},
	} {
		t.Run(fmt.Sprintf("%s/%d", tc.name, tc.lines), func(t *testing.T) {
			if got := functionLengthReviewed(tc.name, tc.lines); got != tc.approved {
				t.Fatalf("reviewed=%v, want %v", got, tc.approved)
			}
		})
	}
}
