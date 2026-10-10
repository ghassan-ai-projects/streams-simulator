package architecture

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// moduleShapes lists the migrated modules (facade over private layers) and
// the internal layers each may have: domain always, app and edge packages
// where the module's kind (STANDARD §1) calls for them. A module that is not
// listed is still a flat package; the program ends when every non-foundation
// module is listed.
var moduleShapes = map[string][]string{
	"internal/perturb": {"domain"},
}

// facadeExemptions names exported facade functions that are deliberately not
// single delegations, with the reason. It is reviewed like any gate table.
var facadeExemptions = map[string]string{}

func migratedModules() []string {
	modules := make([]string, 0, len(moduleShapes))
	for module := range moduleShapes {
		modules = append(modules, module)
	}
	slices.Sort(modules)
	return modules
}

func TestModuleShapeMatchesItsKind(t *testing.T) {
	t.Parallel()
	layers := map[string]map[string]bool{}
	for _, file := range productionFiles(t) {
		module, layer := splitModule(file.pkgDir)
		if layers[module] == nil {
			layers[module] = map[string]bool{}
		}
		if layer != "" {
			layers[module][layer] = true
		}
	}
	for _, module := range migratedModules() {
		allowed := moduleShapes[module]
		if !slices.Contains(allowed, "domain") {
			t.Errorf("%s: a migrated module declares a domain layer", module)
		}
		if !layers[module]["domain"] {
			t.Errorf("%s: declared as migrated but has no internal/domain layer", module)
		}
		for layer := range layers[module] {
			if !slices.Contains(allowed, layer) {
				t.Errorf("%s: internal layer %q is not declared in moduleShapes", module, layer)
			}
		}
	}
}

func TestEveryModuleHasUbiquitousLanguage(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for _, module := range migratedModules() {
		if _, err := root.Stat(module + "/UBIQUITOUS_LANGUAGE.md"); err != nil {
			t.Errorf("%s: missing UBIQUITOUS_LANGUAGE.md", module)
		}
	}
}

// facadeFiles returns the production files of the facade package of module.
func facadeFiles(t *testing.T, module string) []productionFile {
	t.Helper()
	var files []productionFile
	for _, file := range productionFiles(t) {
		if file.pkgDir == module {
			files = append(files, file)
		}
	}
	if len(files) == 0 {
		t.Fatalf("module %s has no facade files", module)
	}
	return files
}

// internalImports returns the local names a file gives the module's internal
// layers.
func internalImports(file productionFile, module string) map[string]bool {
	names := map[string]bool{}
	prefix := modulePrefix + module + "/internal/"
	for name, path := range file.imports {
		if strings.HasPrefix(path, prefix) {
			names[name] = true
		}
	}
	return names
}

func TestFacadesOnlyDelegate(t *testing.T) {
	t.Parallel()
	for _, module := range migratedModules() {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			for _, file := range facadeFiles(t, module) {
				layers := internalImports(file, module)
				for _, decl := range file.source.Decls {
					function, ok := decl.(*ast.FuncDecl)
					if !ok || !function.Name.IsExported() || function.Body == nil {
						continue
					}
					if !delegatesOnly(function, layers) {
						key := file.path + ":" + function.Name.Name
						if _, exempt := facadeExemptions[key]; !exempt {
							t.Errorf("%s: facade function %s does more than delegate", file.path, function.Name.Name)
						}
					}
				}
			}
		})
	}
}

// delegatesOnly accepts a constructor (New…) or a body of exactly one
// statement that calls into an internal layer or through the receiver.
func delegatesOnly(function *ast.FuncDecl, layers map[string]bool) bool {
	if function.Recv == nil && strings.HasPrefix(function.Name.Name, "New") {
		return true
	}
	if len(function.Body.List) != 1 {
		return false
	}
	var expr ast.Expr
	switch statement := function.Body.List[0].(type) {
	case *ast.ReturnStmt:
		if len(statement.Results) != 1 {
			return false
		}
		expr = statement.Results[0]
	case *ast.ExprStmt:
		expr = statement.X
	default:
		return false
	}
	call, ok := expr.(*ast.CallExpr)
	return ok && rootedInLayerOrReceiver(call.Fun, function, layers)
}

func rootedInLayerOrReceiver(fun ast.Expr, function *ast.FuncDecl, layers map[string]bool) bool {
	for {
		selector, ok := fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if ident, ok := selector.X.(*ast.Ident); ok {
			if layers[ident.Name] {
				return true
			}
			return function.Recv != nil && receiverName(function) == ident.Name
		}
		fun = selector.X
	}
}

func receiverName(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 || len(function.Recv.List[0].Names) == 0 {
		return ""
	}
	return function.Recv.List[0].Names[0].Name
}

func TestFacadeSignaturesNameNoInternalTypes(t *testing.T) {
	t.Parallel()
	for _, module := range migratedModules() {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			for _, file := range facadeFiles(t, module) {
				layers := internalImports(file, module)
				for _, leak := range internalTypeLeaks(file, layers) {
					t.Errorf("%s: %s", file.path, leak)
				}
			}
		})
	}
}

// internalTypeLeaks lists exported facade declarations whose type
// expressions name an internal layer. A type alias declaration is the one
// sanctioned way to expose an internal value type.
func internalTypeLeaks(file productionFile, layers map[string]bool) []string {
	var leaks []string
	report := func(what string, node ast.Node) {
		if node != nil && namesLayer(node, layers) {
			leaks = append(leaks, what+" names an internal layer; declare an alias in api.go")
		}
	}
	for _, decl := range file.source.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				report("func "+d.Name.Name, d.Type)
				if d.Recv != nil {
					report("receiver of "+d.Name.Name, d.Recv)
				}
			}
		case *ast.GenDecl:
			leaks = append(leaks, declLeaks(d, layers)...)
		}
	}
	return leaks
}

func declLeaks(decl *ast.GenDecl, layers map[string]bool) []string {
	var leaks []string
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() && s.Assign == token.NoPos && exportedTypeNamesLayer(s.Type, layers) {
				leaks = append(leaks, "type "+s.Name.Name+" exposes an internal layer type; use an alias")
			}
		case *ast.ValueSpec:
			if s.Type != nil && anyExported(s.Names) && namesLayer(s.Type, layers) {
				leaks = append(leaks, "exported value "+s.Names[0].Name+" has an internal layer type")
			}
		}
	}
	return leaks
}

// exportedTypeNamesLayer judges a defined type: for a struct only its
// exported (or embedded) fields are public surface, so an unexported field
// that holds the internal object is not a leak.
func exportedTypeNamesLayer(expr ast.Expr, layers map[string]bool) bool {
	structure, ok := expr.(*ast.StructType)
	if !ok {
		return namesLayer(expr, layers)
	}
	for _, field := range structure.Fields.List {
		if (len(field.Names) == 0 || anyExported(field.Names)) && namesLayer(field.Type, layers) {
			return true
		}
	}
	return false
}

func anyExported(names []*ast.Ident) bool {
	return slices.ContainsFunc(names, func(name *ast.Ident) bool { return name.IsExported() })
}

func namesLayer(node ast.Node, layers map[string]bool) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if selector, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := selector.X.(*ast.Ident); ok && layers[ident.Name] {
				found = true
			}
		}
		return !found
	})
	return found
}
