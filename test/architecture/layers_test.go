package architecture

import (
	"slices"
	"strings"
	"testing"
)

// packageKind is the standard's package kind (STANDARD.md, "Package kinds").
type packageKind string

const (
	kindFoundation packageKind = "foundation" // K1: stdlib and lower foundations only
	kindSeam       packageKind = "seam"       // K1: the one declared clock seam
	kindCore       packageKind = "core"       // K2: pure simulator rules
	kindEdgeCore   packageKind = "edge-core"  // K3: rules plus a named I/O edge
	kindSurface    packageKind = "surface"    // K4: wiring and protocol only
	kindRoot       packageKind = "root"       // module root and development tools
	kindTestAid    packageKind = "test-aid"   // shared test fixtures; tests only
)

// packageInfo is a package's kind and declared layer. An import must point to
// a strictly lower layer, so a new same-layer edge is a reviewed table change
// and not an allowlist line. Layers are the longest dependency path today.
type packageInfo struct {
	kind  packageKind
	layer int
}

// packages classifies every production package directory; a package that is
// not listed fails TestEveryPackageIsClassified.
var packages = map[string]packageInfo{
	"internal/canonical":       {kindFoundation, 0},
	"internal/device/contract": {kindFoundation, 0},
	"internal/randutil":        {kindFoundation, 0},
	"internal/schemas":         {kindFoundation, 0},
	"internal/wall":            {kindSeam, 0},
	"internal/testsupport":     {kindTestAid, 0},
	"internal/sink":            {kindEdgeCore, 0},
	"internal/jsonschema":      {kindFoundation, 1},
	"internal/model":           {kindFoundation, 2},
	"internal/device":          {kindEdgeCore, 2},
	"internal/domain":          {kindEdgeCore, 3},
	"internal/adapter":         {kindEdgeCore, 3},
	"internal/world":           {kindCore, 4},
	"internal/perturb":         {kindCore, 4},
	"internal/run":             {kindEdgeCore, 5},
	"internal/truth":           {kindCore, 5},
	"internal/refconsumer":     {kindCore, 5},
	"internal/deviceworld":     {kindCore, 5},
	"internal/audit":           {kindCore, 5},
	"internal/score":           {kindCore, 5},
	"internal/suite":           {kindCore, 6},
	"internal/mcp":             {kindSurface, 7},
	"internal/cli":             {kindSurface, 8},
	"cmd/streamsim":            {kindSurface, 9},
	".":                        {kindRoot, 10},
	"tools":                    {kindRoot, 10},
}

// splitModule separates a package directory into its module and the name of
// the internal layer inside it ("" for the facade or an unlayered package):
// internal/world/internal/domain is module internal/world, layer domain.
func splitModule(dir string) (module, layer string) {
	rest, ok := strings.CutPrefix(dir, "internal/")
	if !ok {
		return dir, ""
	}
	name, below, layered := strings.Cut(rest, "/internal/")
	if !layered {
		return dir, ""
	}
	layer, _, _ = strings.Cut(below, "/")
	return "internal/" + name, layer
}

// layerRank orders the layers inside one module: domain below edges below app
// below the facade. Inside a module an import must point to a lower rank.
func layerRank(layer string) int {
	switch layer {
	case "":
		return 3
	case "domain":
		return 0
	case "app":
		return 2
	}
	return 1
}

// infoOf classifies any package directory: a module entry, or one of its
// internal layers (domain and app are pure, every other layer is an edge).
func infoOf(dir string) packageInfo {
	module, layer := splitModule(dir)
	info := packages[module]
	switch layer {
	case "":
	case "domain", "app":
		info.kind = kindCore
	default:
		info.kind = kindEdgeCore
	}
	return info
}

func TestEveryPackageIsClassified(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, file := range productionFiles(t) {
		module, _ := splitModule(file.pkgDir)
		seen[module] = true
		if _, ok := packages[module]; !ok {
			t.Errorf("%s: module %s is not classified in packages", file.path, module)
		}
	}
	for dir := range packages {
		if !seen[dir] {
			t.Errorf("stale packages entry: %s has no production files", dir)
		}
	}
}

func TestImportsPointToStrictlyLowerLayers(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		for _, imported := range modulePackages(file) {
			if layerViolation(file.pkgDir, imported) {
				t.Errorf("%s imports %s: imports must point to a lower layer", file.path, imported)
			}
		}
	}
}

func TestFoundationsImportOnlyFoundations(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		if infoOf(file.pkgDir).kind != kindFoundation {
			continue
		}
		for _, imported := range modulePackages(file) {
			if kind := infoOf(imported).kind; kind != kindFoundation {
				t.Errorf("%s: foundation imports %s (%s)", file.path, imported, kind)
			}
		}
	}
}

func TestSurfacesAreImportedOnlyBySurfaces(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		module, _ := splitModule(file.pkgDir)
		if kind := packages[module].kind; kind == kindSurface || kind == kindRoot {
			continue
		}
		for _, imported := range modulePackages(file) {
			if infoOf(imported).kind == kindSurface {
				t.Errorf("%s imports surface package %s", file.path, imported)
			}
		}
	}
}

// modulePackages returns the sorted, de-duplicated package directories of this
// module that a file imports.
func modulePackages(file productionFile) []string {
	var packages []string
	for _, path := range file.paths {
		if directory, ok := strings.CutPrefix(path, modulePrefix); ok {
			packages = append(packages, directory)
		}
	}
	slices.Sort(packages)
	return slices.Compact(packages)
}

// layerViolation reports an import that does not point to a strictly lower
// layer: between modules by module layer, inside a module by layer rank.
func layerViolation(owner, imported string) bool {
	ownerModule, ownerLayer := splitModule(owner)
	importedModule, importedLayer := splitModule(imported)
	if ownerModule == importedModule {
		return layerRank(ownerLayer) <= layerRank(importedLayer)
	}
	return packages[ownerModule].layer <= packages[importedModule].layer
}

func TestLayerViolationRejectsSameAndUpwardEdges(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		owner, imported string
		violation       bool
	}{
		{"internal/world", "internal/domain", false},
		{"internal/world", "internal/perturb", true}, // same layer
		{"internal/model", "internal/world", true},   // upward
	} {
		if got := layerViolation(tc.owner, tc.imported); got != tc.violation {
			t.Errorf("%s -> %s: violation=%v, want %v", tc.owner, tc.imported, got, tc.violation)
		}
	}
}

// TestAllowedImportsHaveNoStaleEdges keeps the direct-import allowlist honest:
// an approved edge that no production file uses is removed.
func TestAllowedImportsHaveNoStaleEdges(t *testing.T) {
	t.Parallel()
	used := map[string]map[string]bool{}
	for _, file := range productionFiles(t) {
		owner, _ := splitModule(file.pkgDir)
		if used[owner] == nil {
			used[owner] = map[string]bool{}
		}
		for _, imported := range modulePackages(file) {
			if module, _ := splitModule(imported); module != owner {
				used[owner][strings.TrimPrefix(module, "internal/")] = true
			}
		}
	}
	for owner, allowed := range packageDependencies {
		for _, dependency := range strings.Fields(allowed) {
			if !used[owner][dependency] {
				t.Errorf("stale allowlist edge: %s -> %s is imported by no production file", owner, dependency)
			}
		}
	}
}
