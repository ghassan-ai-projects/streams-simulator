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

// maxFunctionBodyLines is the production function bound (STANDARD M8): the
// physical lines from opening to closing brace, comments and blanks included.
const maxFunctionBodyLines = 15

func TestProductionFunctionsStayWithinTheBodyLimit(t *testing.T) {
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
			lines := positions.Position(fn.Body.End()).Line - positions.Position(fn.Body.Pos()).Line + 1
			if lines > maxFunctionBodyLines {
				t.Errorf("%s has a %d-line body; maximum is %d: extract named steps", key, lines, maxFunctionBodyLines)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncLit)
			if !ok {
				return true
			}
			start := positions.Position(fn.Body.Pos())
			lines := positions.Position(fn.Body.End()).Line - start.Line + 1
			if lines > maxFunctionBodyLines {
				t.Errorf("%s:%d has a %d-line anonymous function: extract a named operation", path, start.Line, lines)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
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
