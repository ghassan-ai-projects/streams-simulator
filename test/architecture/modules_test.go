package architecture

import (
	"go/ast"
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
	"internal/perturb":     {"domain"},
	"internal/truth":       {"domain"},
	"internal/world":       {"domain"},
	"internal/score":       {"domain"},
	"internal/audit":       {"domain"},
	"internal/suite":       {"domain"},
	"internal/deviceworld": {"domain"},
	"internal/sink":        {"domain", "files", "httppush"},
	"internal/domain":      {"domain", "files"},
	"internal/adapter":     {"domain", "files"},
	"internal/device":      {"domain", "uds"},
	"internal/run":         {"domain", "app", "durable", "quiesce", "clock"},
	"internal/refconsumer": {"domain", "mcpclient"},
}

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

// layerMethods lists the sorted exported methods of a type declared in the
// production package at the import path; an alias of the type exports them.
func layerMethods(t *testing.T) func(path, typeName string) []string {
	t.Helper()
	return func(path, typeName string) []string {
		directory := strings.TrimPrefix(path, modulePrefix)
		var methods []string
		for _, file := range productionFiles(t) {
			if file.pkgDir != directory {
				continue
			}
			for _, decl := range file.source.Decls {
				if function, ok := decl.(*ast.FuncDecl); ok && function.Name.IsExported() && receiverType(function) == typeName {
					methods = append(methods, function.Name.Name)
				}
			}
		}
		slices.Sort(methods)
		return methods
	}
}

// TestAliasReviewsAreAllInUse keeps aliasedValueTypes honest: an entry whose
// alias no facade declares any more is removed.
func TestAliasReviewsAreAllInUse(t *testing.T) {
	t.Parallel()
	declared := map[string]bool{}
	for _, file := range productionFiles(t) {
		for _, decl := range file.source.Decls {
			general, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range general.Specs {
				if alias, ok := spec.(*ast.TypeSpec); ok && alias.Assign.IsValid() {
					declared[file.pkgDir+":"+alias.Name.Name] = true
				}
			}
		}
	}
	for key := range aliasedValueTypes {
		if !declared[key] {
			t.Errorf("aliasedValueTypes has %s but no facade declares that alias", key)
		}
	}
}

func TestFacadesFollowTheFacadeRules(t *testing.T) {
	t.Parallel()
	methodsOf := layerMethods(t)
	for _, module := range migratedModules() {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			files := facadeFiles(t, module)
			for _, file := range files {
				for _, violation := range facadeViolations(file, module, methodsOf, files...) {
					t.Errorf("%s: %s", file.path, violation)
				}
			}
		})
	}
}
