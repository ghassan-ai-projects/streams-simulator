//go:build ignore

// Run this source checker with go run cmd/streamsim/function_length.go.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const maximumFunctionBodyLines = 15

func main() {
	violations := 0
	err := filepath.WalkDir(".", functionFileVisitor(&violations))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if violations != 0 {
		fmt.Fprintf(os.Stderr, "%d production functions exceed %d body lines\n", violations, maximumFunctionBodyLines)
		os.Exit(1)
	}
	fmt.Println("production function length check passed")
}

func functionFileVisitor(violations *int) fs.WalkDirFunc {
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return fs.SkipDir
		}
		if !isProductionGoSource(path, entry) {
			return nil
		}
		return inspectProductionFunctions(path, violations)
	}
}

func isProductionGoSource(path string, entry fs.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func inspectProductionFunctions(path string, violations *int) error {
	positions := token.NewFileSet()
	source, err := parser.ParseFile(positions, path, nil, 0)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	ast.Inspect(source, func(node ast.Node) bool {
		if body := functionBody(node); body != nil {
			reportFunctionLength(positions, body, violations)
		}
		return true
	})
	return nil
}

func functionBody(node ast.Node) *ast.BlockStmt {
	switch fn := node.(type) {
	case *ast.FuncDecl:
		return fn.Body
	case *ast.FuncLit:
		return fn.Body
	default:
		return nil
	}
}

func reportFunctionLength(positions *token.FileSet, body *ast.BlockStmt, violations *int) {
	start := positions.Position(body.Pos())
	lines := positions.Position(body.End()).Line - start.Line + 1
	if lines > maximumFunctionBodyLines {
		fmt.Printf("%s:%d: %d body lines; maximum %d\n", start.Filename, start.Line, lines, maximumFunctionBodyLines)
		*violations++
	}
}
