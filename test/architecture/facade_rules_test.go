package architecture

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

// facadeExemptions names facade functions that deliberately do more than
// delegate, keyed "<file path>:<function>", with the reason. It is a
// reviewed table: an entry is a design decision, not a convenience.
var facadeExemptions = map[string]string{}

// aliasReview pins a reviewed alias of a method-bearing layer type: the layer
// type it names and the exact exported method set it exposes. A new method on
// the layer type fails the gate until the review is renewed here.
type aliasReview struct {
	target  string
	methods []string
	reason  string
}

// aliasedValueTypes lists facade aliases of layer types that have methods,
// keyed "<module>:<alias>". Each is a value type whose method set is read-only
// lookup over its own data and is the contract other modules use; its
// exported fields (for Compiled, Spec/Digest/Raw) are deliberately plain data.
var aliasedValueTypes = map[string]aliasReview{
	"internal/domain:Compiled": {
		target: "Compiled",
		methods: []string{"Channel", "ChannelGain", "ChannelNames", "Effector", "Fault", "HasChannel",
			"HasEffector", "HasFault", "HasProfile", "HasState", "Profile", "StateNames"},
		reason: "compiled domain spec: lookups over its own declared names; Spec is deliberately mutable for tests",
	},
	"internal/device:Capabilities": {
		target:  "Capabilities",
		methods: []string{"Digest", "SafeStopNames"},
		reason:  "device capability catalog: read-only target and safe-stop lookups plus its digest",
	},
}

// facade describes one facade file for the rules below.
type facade struct {
	module    string
	file      productionFile
	layers    map[string]string          // local name -> import path of an internal layer
	holders   map[string]map[string]bool // struct type -> field -> holds an internal layer object
	methodsOf func(importPath, typeName string) []string
}

// facadeViolations returns every way a facade file breaks the module rules:
// it delegates only, exposes no internal layer type and re-exports no mutable
// state (STANDARD M5, M6). siblings are the module's other facade files,
// whose struct declarations say which fields hold layer objects.
func facadeViolations(file productionFile, module string, methodsOf func(string, string) []string, siblings ...productionFile) []string {
	f := facade{module: module, file: file, layers: map[string]string{}, holders: map[string]map[string]bool{}, methodsOf: methodsOf}
	violations := f.collectLayers(module)
	for _, peer := range append([]productionFile{file}, siblings...) {
		peerView := facade{file: peer, layers: map[string]string{}, holders: f.holders}
		peerView.collectLayers(module)
		peerView.collectHolders()
	}
	for _, decl := range file.source.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			violations = append(violations, f.functionViolations(d)...)
		case *ast.GenDecl:
			violations = append(violations, f.declarationViolations(d)...)
		}
	}
	return violations
}

func (f *facade) collectLayers(module string) []string {
	var violations []string
	prefix := modulePrefix + module + "/internal/"
	for _, spec := range f.file.source.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "." || name == "_" {
			violations = append(violations, "the internal layer "+path+" is imported with a dot or blank name")
			continue
		}
		f.layers[name] = path
	}
	return violations
}

func (f *facade) collectHolders() {
	for _, decl := range f.file.source.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				f.recordHolder(typeSpec)
			}
		}
	}
}

func (f *facade) recordHolder(spec *ast.TypeSpec) {
	structure, ok := spec.Type.(*ast.StructType)
	if !ok {
		return
	}
	fields := map[string]bool{}
	for _, field := range structure.Fields.List {
		for _, name := range field.Names {
			fields[name.Name] = f.namesLayer(field.Type)
		}
	}
	f.holders[spec.Name.Name] = fields
}

func (f *facade) functionViolations(function *ast.FuncDecl) []string {
	key := f.file.path + ":" + function.Name.Name
	if _, exempt := facadeExemptions[key]; exempt || function.Body == nil {
		return nil
	}
	var violations []string
	if function.Name.Name == "init" {
		violations = append(violations, "init is not allowed in a facade")
	}
	if function.Name.IsExported() && f.namesLayer(function.Type) {
		violations = append(violations, "func "+function.Name.Name+" names an internal layer in its signature; declare an alias in api.go")
	}
	if function.Name.IsExported() && function.Recv != nil && receiverNamesLayer(f, function) {
		violations = append(violations, "method "+function.Name.Name+" has an internal layer receiver")
	}
	if !f.delegates(function) {
		violations = append(violations, "function "+function.Name.Name+" does more than delegate")
	}
	return violations
}

func receiverNamesLayer(f *facade, function *ast.FuncDecl) bool {
	return function.Recv != nil && f.namesLayer(function.Recv)
}

// delegates accepts a constructor (guards, then one return) or a body of
// exactly one statement that calls into a layer, or through a receiver field
// that holds a layer object, with plain arguments.
func (f *facade) delegates(function *ast.FuncDecl) bool {
	if function.Recv == nil && strings.HasPrefix(function.Name.Name, "New") {
		return f.constructs(function)
	}
	call, ok := singleCall(f.withoutGuards(function.Body.List))
	return ok && f.calleeIsLayer(call.Fun, function) && plainArguments(call)
}

// withoutGuards drops the leading nil guards (a guard returns nil, zero
// values or an Err… sentinel) so a delegation may refuse a missing handle
// before it calls through.
func (f *facade) withoutGuards(list []ast.Stmt) []ast.Stmt {
	for len(list) > 1 {
		guard, ok := list[0].(*ast.IfStmt)
		if !ok || !f.constructorStep(guard) {
			break
		}
		list = list[1:]
	}
	return list
}

func singleCall(list []ast.Stmt) (*ast.CallExpr, bool) {
	if len(list) != 1 {
		return nil, false
	}
	var expr ast.Expr
	switch statement := list[0].(type) {
	case *ast.ReturnStmt:
		if len(statement.Results) != 1 {
			return nil, false
		}
		expr = statement.Results[0]
	case *ast.ExprStmt:
		expr = statement.X
	default:
		return nil, false
	}
	call, ok := expr.(*ast.CallExpr)
	return call, ok
}

// calleeIsLayer requires `<layer>.F(...)` or `<receiver>.<field>.M(...)`
// where the field was declared with an internal layer type.
func (f *facade) calleeIsLayer(fun ast.Expr, function *ast.FuncDecl) bool {
	selector, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch root := selector.X.(type) {
	case *ast.Ident:
		_, isLayer := f.layers[root.Name]
		return isLayer
	case *ast.SelectorExpr:
		ident, ok := root.X.(*ast.Ident)
		return ok && function.Recv != nil && ident.Name == receiverName(function) &&
			f.holders[receiverType(function)][root.Sel.Name]
	}
	return false
}

// plainArguments rejects arguments that hide logic: function literals and
// nested calls other than conversions to a predeclared type.
func plainArguments(call *ast.CallExpr) bool {
	for _, argument := range call.Args {
		plain := true
		ast.Inspect(argument, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncLit, *ast.BinaryExpr, *ast.IndexExpr, *ast.SliceExpr, *ast.StarExpr:
				plain = false
			case *ast.CallExpr:
				if !isConversion(n) {
					plain = false
				}
			}
			return plain
		})
		if !plain {
			return false
		}
	}
	return true
}

func isConversion(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && token.IsKeyword(ident.Name) == false && slices.Contains(predeclaredTypes, ident.Name)
}

var predeclaredTypes = []string{
	"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64",
	"float32", "float64", "string", "byte", "rune", "bool", "any",
}
