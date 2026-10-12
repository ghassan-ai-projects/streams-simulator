package architecture

import (
	"go/ast"
	"go/token"
	"strconv"
	"testing"
)

// TestSkipUnderRaceOnlyMarksSingleGoroutineTests keeps the race tier honest:
// testsupport.SkipUnderRace removes a test from the -race pass, so it may only
// mark a test that starts no goroutine and it must say why.
func TestSkipUnderRaceOnlyMarksSingleGoroutineTests(t *testing.T) {
	t.Parallel()
	for _, file := range testFiles(t) {
		for _, decl := range file.source.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			checkSkipUnderRace(t, file, function)
		}
	}
}

func checkSkipUnderRace(t *testing.T, file productionFile, function *ast.FuncDecl) {
	t.Helper()
	var skips []*ast.CallExpr
	starts := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.GoStmt:
			starts = true
		case *ast.CallExpr:
			if isSkipUnderRace(n) {
				skips = append(skips, n)
			}
		}
		return true
	})
	for _, call := range skips {
		if starts {
			t.Errorf("%s: %s starts a goroutine, so it must stay under -race", file.path, function.Name.Name)
		}
		if !hasReason(call) {
			t.Errorf("%s: %s calls SkipUnderRace without a reason naming what it computes", file.path, function.Name.Name)
		}
	}
}

func isSkipUnderRace(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "SkipUnderRace"
}

func hasReason(call *ast.CallExpr) bool {
	if len(call.Args) != 2 {
		return false
	}
	literal, ok := call.Args[1].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	reason, err := strconv.Unquote(literal.Value)
	return err == nil && len(reason) >= 10
}
