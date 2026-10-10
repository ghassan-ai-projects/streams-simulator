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

// layerHasMethods reports whether a type declared in the production package
// at the import path has methods; an alias of such a type would export them.
func layerHasMethods(t *testing.T) func(path, typeName string) bool {
	t.Helper()
	return func(path, typeName string) bool {
		directory := strings.TrimPrefix(path, modulePrefix)
		for _, file := range productionFiles(t) {
			if file.pkgDir != directory {
				continue
			}
			for _, decl := range file.source.Decls {
				if function, ok := decl.(*ast.FuncDecl); ok && function.Name.IsExported() && receiverType(function) == typeName {
					return true
				}
			}
		}
		return false
	}
}

func TestFacadesFollowTheFacadeRules(t *testing.T) {
	t.Parallel()
	hasMethods := layerHasMethods(t)
	for _, module := range migratedModules() {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			files := facadeFiles(t, module)
			for _, file := range files {
				for _, violation := range facadeViolations(file, module, hasMethods, files...) {
					t.Errorf("%s: %s", file.path, violation)
				}
			}
		})
	}
}
