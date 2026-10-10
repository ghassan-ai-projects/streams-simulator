package architecture

import (
	"go/ast"
	"go/token"
	"strings"
)

// constructs accepts the shape of a fail-closed constructor: nil guards that
// return, definitions from a single layer call, and one final return of a
// call or composite literal. No loops, no function literals.
func (f *facade) constructs(function *ast.FuncDecl) bool {
	list := function.Body.List
	if len(list) == 0 {
		return false
	}
	definitions := 0
	for _, statement := range list[:len(list)-1] {
		if _, defines := statement.(*ast.AssignStmt); defines {
			definitions++
		}
		if !f.constructorStep(statement) || definitions > 1 {
			return false
		}
	}
	returned, ok := list[len(list)-1].(*ast.ReturnStmt)
	return ok && len(returned.Results) >= 1 && f.plainResults(returned.Results)
}

// plainResults requires every returned value to be a nil/Err value, a plain
// reference, a layer constructor call with plain arguments, or a composite
// literal whose field values are plain.
func (f *facade) plainResults(results []ast.Expr) bool {
	for _, result := range results {
		if !f.plainValue(result) {
			return false
		}
	}
	return true
}

func (f *facade) plainValue(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident, *ast.BasicLit:
		return true
	case *ast.SelectorExpr:
		return f.plainValue(e.X)
	case *ast.StarExpr:
		return f.plainValue(e.X)
	case *ast.UnaryExpr:
		return e.Op == token.AND && f.plainValue(e.X)
	case *ast.CompositeLit:
		return f.plainFields(e.Elts)
	case *ast.CallExpr:
		return (f.calleeIsLayerFunction(e.Fun) || isConversion(e)) && f.plainArgumentValues(e.Args)
	}
	return false
}

func (f *facade) plainFields(elements []ast.Expr) bool {
	for _, element := range elements {
		if pair, ok := element.(*ast.KeyValueExpr); ok {
			element = pair.Value
		}
		if !f.plainValue(element) {
			return false
		}
	}
	return true
}

func (f *facade) plainArgumentValues(arguments []ast.Expr) bool {
	for _, argument := range arguments {
		if !f.plainValue(argument) {
			return false
		}
	}
	return true
}

// constructorStep judges one statement before the final return: a refusing
// nil guard or a definition from a single layer call.
func (f *facade) constructorStep(statement ast.Stmt) bool {
	switch s := statement.(type) {
	case *ast.IfStmt:
		return s.Init == nil && s.Else == nil && isNilComparison(s.Cond) && f.guardReturns(s.Body.List)
	case *ast.AssignStmt:
		call, ok := singleAssignedCall(s)
		return ok && s.Tok == token.DEFINE && f.calleeIsLayerFunction(call.Fun) && plainArguments(call)
	}
	return false
}

// guardReturns accepts a guard body that is one return of nil, zero values
// and Err… sentinels only: a guard refuses, it never repairs.
func (f *facade) guardReturns(body []ast.Stmt) bool {
	if len(body) != 1 {
		return false
	}
	returned, ok := body[0].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	for _, result := range returned.Results {
		if !refusalValue(result) {
			return false
		}
	}
	return true
}

func refusalValue(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "nil" || e.Name == "false" || e.Name == "err" || strings.HasPrefix(e.Name, "Err")
	case *ast.BasicLit:
		return e.Value == "0" || e.Value == `""`
	case *ast.SelectorExpr:
		return strings.HasPrefix(e.Sel.Name, "Err")
	}
	return false
}

func singleAssignedCall(assignment *ast.AssignStmt) (*ast.CallExpr, bool) {
	if len(assignment.Rhs) != 1 {
		return nil, false
	}
	call, ok := assignment.Rhs[0].(*ast.CallExpr)
	return call, ok
}

func (f *facade) calleeIsLayerFunction(fun ast.Expr) bool {
	selector, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, isLayer := f.layers[ident.Name]
	return isLayer
}

func isNilComparison(expr ast.Expr) bool {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if binary.Op == token.LOR || binary.Op == token.LAND {
		return isNilComparison(binary.X) && isNilComparison(binary.Y)
	}
	nilSide := func(e ast.Expr) bool { ident, ok := e.(*ast.Ident); return ok && ident.Name == "nil" }
	if binary.Op != token.EQL && binary.Op != token.NEQ {
		return false
	}
	return (nilSide(binary.X) && referenceChain(binary.Y)) || (nilSide(binary.Y) && referenceChain(binary.X))
}

// referenceChain accepts an identifier or a selector chain over identifiers:
// the operand of a nil guard is a value, never a call or an index.
func referenceChain(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return referenceChain(e.X)
	}
	return false
}
