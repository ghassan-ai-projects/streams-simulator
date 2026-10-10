package architecture

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

func (f *facade) declarationViolations(decl *ast.GenDecl) []string {
	var violations []string
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			violations = append(violations, f.typeViolations(s)...)
		case *ast.ValueSpec:
			violations = append(violations, f.valueViolations(decl.Tok, s)...)
		}
	}
	return violations
}

func (f *facade) typeViolations(spec *ast.TypeSpec) []string {
	if spec.Assign != token.NoPos {
		return f.aliasViolations(spec)
	}
	if spec.Name.IsExported() && f.exportedTypeNamesLayer(spec.Type) {
		return []string{"type " + spec.Name.Name + " exposes an internal layer type; use an alias of a plain value record"}
	}
	return nil
}

// aliasViolations allows an alias of a layer type only for an exported plain
// value record declared in api.go: an alias of a type with methods would
// export its whole method set past the facade.
func (f *facade) aliasViolations(spec *ast.TypeSpec) []string {
	if !f.namesLayer(spec.Type) {
		return nil
	}
	selector, plain := spec.Type.(*ast.SelectorExpr)
	if !plain {
		return []string{"alias " + spec.Name.Name + " is a composite over an internal layer type; alias the plain type"}
	}
	if !spec.Name.IsExported() || filepath.Base(f.file.path) != "api.go" {
		return []string{"alias " + spec.Name.Name + " of an internal layer type must be exported and declared in api.go"}
	}
	ident := selector.X.(*ast.Ident)
	methods := f.methodsOf(f.layers[ident.Name], selector.Sel.Name)
	if len(methods) == 0 {
		return nil
	}
	review, reviewed := aliasedValueTypes[f.module+":"+spec.Name.Name]
	if !reviewed || review.target != selector.Sel.Name {
		return []string{"alias " + spec.Name.Name + " would export the methods of " + selector.Sel.Name + "; wrap the type in a facade struct"}
	}
	if !slices.Equal(methods, review.methods) {
		return []string{"alias " + spec.Name.Name + " exposes " + strings.Join(methods, ",") + ", reviewed " + strings.Join(review.methods, ",")}
	}
	return nil
}

func (f *facade) valueViolations(tok token.Token, spec *ast.ValueSpec) []string {
	var violations []string
	for _, name := range spec.Names {
		if !name.IsExported() || tok == token.CONST || strings.HasPrefix(name.Name, "Err") {
			continue
		}
		violations = append(violations, "exported variable "+name.Name+" is mutable state past the facade; expose a function")
	}
	return violations
}

func (f *facade) exportedTypeNamesLayer(expr ast.Expr) bool {
	structure, ok := expr.(*ast.StructType)
	if !ok {
		return f.namesLayer(expr)
	}
	for _, field := range structure.Fields.List {
		if (len(field.Names) == 0 || anyExported(field.Names)) && f.namesLayer(field.Type) {
			return true
		}
	}
	return false
}

func (f *facade) namesLayer(node ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if selector, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := selector.X.(*ast.Ident); ok {
				if _, isLayer := f.layers[ident.Name]; isLayer {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

func anyExported(names []*ast.Ident) bool {
	return slices.ContainsFunc(names, func(name *ast.Ident) bool { return name.IsExported() })
}

func receiverName(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 || len(function.Recv.List[0].Names) == 0 {
		return ""
	}
	return function.Recv.List[0].Names[0].Name
}

// receiverType is the base type name of a method receiver ("" for functions).
func receiverType(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return ""
	}
	expr := function.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
