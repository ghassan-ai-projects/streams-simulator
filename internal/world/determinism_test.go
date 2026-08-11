package world

// Determinism guard: no output-producing code may iterate a map to produce
// output, because map iteration order is unspecified. This test scans the
// world and adapter packages: it collects every map-typed field and
// variable and fails on any `range` over one in non-test code. It is a lint
// for a rule Go cannot express.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoMapIterationInOutputCode(t *testing.T) {
	roots := []string{".", "../adapter", "../perturb", "../ledger", "../run", "../truth", "../audit", "../score", "../refconsumer", "../mcp", "../domain", "../sink", "../suite", "../canonical", "../randutil", "../jsonschema", "../model", "../schemas", "../wall"}
	checked := 0
	for _, root := range roots {
		files, err := filepath.Glob(filepath.Join(root, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		checked += len(files)
		fields, vars := collectMapNames(files, t)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, f, src, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(file, func(n ast.Node) bool {
				rs, ok := n.(*ast.RangeStmt)
				if !ok {
					return true
				}
				pos := fset.Position(rs.Pos())
				line := string(src[rs.Pos()-1 : rs.End()-1])
				_ = line
				if rangesMap(rs.X, fields, vars) && !markerOnLine(src, pos.Line) {
					t.Errorf("%s:%d: ranging over a map in output code is nondeterministic", pos.Filename, pos.Line)
				}
				return true
			})
		}
	}
	if checked == 0 {
		t.Fatal("no files scanned")
	}
}

// collectMapNames returns (map-typed struct field names, map-typed local
// variable names) across the given files.
func collectMapNames(files []string, t *testing.T) (map[string]bool, map[string]bool) {
	fields := map[string]bool{}
	vars := map[string]bool{}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if st, ok := x.Type.(*ast.StructType); ok {
					for _, f := range st.Fields.List {
						if _, ok := f.Type.(*ast.MapType); ok {
							for _, name := range f.Names {
								fields[name.Name] = true
							}
						}
					}
				}
			case *ast.ValueSpec:
				for i, name := range x.Names {
					if len(x.Values) > i {
						if _, ok := x.Values[i].(*ast.MapType); ok {
							vars[name.Name] = true
						}
					}
					if _, ok := x.Type.(*ast.MapType); ok {
						vars[name.Name] = true
					}
				}
			case *ast.AssignStmt:
				for i, name := range x.Lhs {
					if id, ok := name.(*ast.Ident); ok && len(x.Rhs) > i {
						if _, ok := x.Rhs[i].(*ast.MapType); ok {
							vars[id.Name] = true
						}
						if c, ok := x.Rhs[i].(*ast.CallExpr); ok {
							if fn, ok := c.Fun.(*ast.Ident); ok && fn.Name == "make" {
								// Only map-typed make() calls.
								if len(c.Args) > 0 {
									if _, ok := c.Args[0].(*ast.MapType); ok {
										vars[id.Name] = true
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return fields, vars
}

// markerOnLine reports whether a marker appears within a small window
// around the range statement (the comment may sit just above the loop).
func markerOnLine(src []byte, line int) bool {
	lines := strings.Split(string(src), "\n")
	for i := line - 2; i <= line; i++ {
		if i >= 0 && i < len(lines) && strings.Contains(lines[i], "determinism-safe") {
			return true
		}
	}
	return false
}

// rangesMap reports whether the ranged expression touches a known map.
// Indexing a map yields its value (usually a slice), which is safe to
// range; only ranging the map itself is flagged.
func rangesMap(expr ast.Expr, fields, vars map[string]bool) bool {
	switch x := expr.(type) {
	case *ast.MapType:
		return true
	case *ast.Ident:
		return vars[x.Name]
	case *ast.SelectorExpr:
		return fields[x.Sel.Name]
	case *ast.CallExpr:
		return rangesMap(x.Fun, fields, vars)
	}
	return false
}
