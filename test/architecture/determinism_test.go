package architecture

// Determinism guard: no output-producing code may range over a map, because
// map iteration order is unspecified. The gate collects every map-typed
// field and variable of a package (all its files, layers included) and fails
// on a `range` over one in production code unless a "determinism-safe"
// comment sits on or just above the loop. It lints a rule Go cannot express.

import (
	"go/ast"
	"strings"
	"testing"
)

const determinismMarker = "determinism-safe"

// determinismDebt lists packages the map-iteration gate does not yet cover
// (it covered the world, perturbation, run and scoring packages before this
// program). Their map ranges only order which error message or catalog entry
// surfaces first (DEFERRED D-38), never emitted records; each entry leaves
// when its module is migrated and the ranges are made ordered.
var determinismDebt = map[string]string{
	"internal/device":                      "capability and fault validation order (D-38); module M11",
	"internal/deviceworld/internal/domain": "binding validation order (D-38); module M7 (layer)",
}

func TestNoMapIterationInOutputCode(t *testing.T) {
	t.Parallel()
	byPackage := map[string][]productionFile{}
	for _, file := range productionFiles(t) {
		byPackage[file.pkgDir] = append(byPackage[file.pkgDir], file)
	}
	for directory, files := range byPackage {
		if _, debt := determinismDebt[directory]; debt {
			continue
		}
		fields, vars := mapNames(files)
		for _, file := range files {
			markers := markerLines(file)
			ast.Inspect(file.source, func(node ast.Node) bool {
				statement, ok := node.(*ast.RangeStmt)
				if !ok || !rangesMap(statement.X, fields, vars) {
					return true
				}
				line := file.fset.Position(statement.Pos()).Line
				if !markers[line-1] && !markers[line] && !markers[line+1] {
					t.Errorf("%s:%d (%s): ranging over a map in output code is nondeterministic", file.path, line, directory)
				}
				return true
			})
		}
	}
}

// markerLines returns the lines a determinism marker comment covers.
func markerLines(file productionFile) map[int]bool {
	lines := map[int]bool{}
	for _, group := range file.source.Comments {
		for _, comment := range group.List {
			if strings.Contains(comment.Text, determinismMarker) {
				start := file.fset.Position(comment.Pos()).Line
				end := file.fset.Position(comment.End()).Line
				for line := start; line <= end; line++ {
					lines[line] = true
				}
			}
		}
	}
	return lines
}

// mapNames returns the names of map-typed struct fields and of map-typed
// variables across the files of one package.
func mapNames(files []productionFile) (fields, vars map[string]bool) {
	fields, vars = map[string]bool{}, map[string]bool{}
	for _, file := range files {
		ast.Inspect(file.source, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.TypeSpec:
				recordMapFields(x, fields)
			case *ast.ValueSpec:
				recordMapValues(x, vars)
			case *ast.AssignStmt:
				recordMapAssignments(x, vars)
			}
			return true
		})
	}
	return fields, vars
}

func recordMapFields(spec *ast.TypeSpec, fields map[string]bool) {
	structure, ok := spec.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, field := range structure.Fields.List {
		if _, isMap := field.Type.(*ast.MapType); isMap {
			for _, name := range field.Names {
				fields[name.Name] = true
			}
		}
	}
}

func recordMapValues(spec *ast.ValueSpec, vars map[string]bool) {
	for i, name := range spec.Names {
		if len(spec.Values) > i {
			if _, isMap := spec.Values[i].(*ast.MapType); isMap {
				vars[name.Name] = true
			}
		}
		if _, isMap := spec.Type.(*ast.MapType); isMap {
			vars[name.Name] = true
		}
	}
}

func recordMapAssignments(assignment *ast.AssignStmt, vars map[string]bool) {
	for i, target := range assignment.Lhs {
		ident, ok := target.(*ast.Ident)
		if !ok || len(assignment.Rhs) <= i {
			continue
		}
		switch rhs := assignment.Rhs[i].(type) {
		case *ast.MapType:
			vars[ident.Name] = true
		case *ast.CallExpr:
			if makesMap(rhs) {
				vars[ident.Name] = true
			}
		}
	}
}

func makesMap(call *ast.CallExpr) bool {
	function, ok := call.Fun.(*ast.Ident)
	if !ok || function.Name != "make" || len(call.Args) == 0 {
		return false
	}
	_, isMap := call.Args[0].(*ast.MapType)
	return isMap
}

// rangesMap reports whether the ranged expression is a known map. Indexing a
// map yields its value (usually a slice), which is safe to range.
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
