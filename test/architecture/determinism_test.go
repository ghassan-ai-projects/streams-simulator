package architecture

// Determinism guard: no output-producing code may range over a map, because
// map iteration order is unspecified. The gate type-checks every production
// package and fails on a `range` whose operand's type is a map unless a
// "determinism-safe" comment states why the order cannot reach an output
// (the keys are sorted afterwards, the loop only copies or aggregates, ...).
// It lints a rule Go cannot express, using types, not names.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"testing"

	xpackages "golang.org/x/tools/go/packages"
)

const determinismMarker = "determinism-safe"

func TestNoMapIterationInOutputCode(t *testing.T) {
	t.Parallel()
	loaded, err := xpackages.Load(&xpackages.Config{
		Mode: xpackages.NeedName | xpackages.NeedFiles | xpackages.NeedSyntax | xpackages.NeedTypes |
			xpackages.NeedTypesInfo | xpackages.NeedImports | xpackages.NeedDeps,
		Dir: "../..",
	}, "./internal/...", "./cmd/...")
	if err != nil {
		t.Fatalf("type-check the module: %v", err)
	}
	for _, pkg := range loaded {
		for _, problem := range pkg.Errors {
			t.Fatalf("package %s does not type-check: %v", pkg.PkgPath, problem)
		}
		for _, file := range pkg.Syntax {
			checkMapRanges(t, pkg, file)
		}
	}
}

func checkMapRanges(t *testing.T, pkg *xpackages.Package, file *ast.File) {
	t.Helper()
	markers := markerLines(pkg.Fset, file)
	ast.Inspect(file, func(node ast.Node) bool {
		statement, ok := node.(*ast.RangeStmt)
		if !ok || !isMapType(pkg.TypesInfo.TypeOf(statement.X)) {
			return true
		}
		position := pkg.Fset.Position(statement.Pos())
		if !markers[position.Line-1] && !markers[position.Line] && !markers[position.Line+1] {
			t.Errorf("%s:%d: ranging over a map is order-dependent; sort the keys or add a %q comment saying why the order cannot reach output",
				strings.TrimPrefix(position.Filename, repositoryPrefix(position.Filename)), position.Line, determinismMarker)
		}
		return true
	})
}

func isMapType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, isMap := t.Underlying().(*types.Map)
	return isMap
}

// markerLines returns the lines a determinism marker comment covers.
func markerLines(fset *token.FileSet, file *ast.File) map[int]bool {
	lines := map[int]bool{}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.Contains(comment.Text, determinismMarker) {
				for line := fset.Position(comment.Pos()).Line; line <= fset.Position(comment.End()).Line; line++ {
					lines[line] = true
				}
			}
		}
	}
	return lines
}

// repositoryPrefix returns the absolute path up to and including the first
// "internal/" or "cmd/" element, so reports show repository-relative paths.
func repositoryPrefix(filename string) string {
	for _, element := range []string{"/internal/", "/cmd/"} {
		if index := strings.Index(filename, element); index >= 0 {
			return filename[:index+1]
		}
	}
	return ""
}
