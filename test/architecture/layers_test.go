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
	"internal/canonical":           {kindFoundation, 0},
	"internal/randutil":            {kindFoundation, 0},
	"internal/schemas":             {kindFoundation, 0},
	"internal/wall":                {kindSeam, 0},
	"internal/sink":                {kindEdgeCore, 0},
	"internal/jsonschema":          {kindFoundation, 1},
	"internal/model":               {kindFoundation, 2},
	"internal/device":              {kindEdgeCore, 2},
	"internal/domain":              {kindEdgeCore, 3},
	"internal/adapter":             {kindEdgeCore, 3},
	"internal/adapter/conformance": {kindEdgeCore, 4},
	"internal/world":               {kindCore, 4},
	"internal/perturb":             {kindCore, 4},
	"internal/run":                 {kindEdgeCore, 5},
	"internal/truth":               {kindCore, 5},
	"internal/refconsumer":         {kindCore, 5},
	"internal/deviceworld":         {kindCore, 5},
	"internal/audit":               {kindCore, 6},
	"internal/score":               {kindCore, 6},
	"internal/suite":               {kindCore, 7},
	"internal/mcp":                 {kindSurface, 7},
	"internal/cli":                 {kindSurface, 8},
	"cmd/streamsim":                {kindSurface, 9},
	".":                            {kindRoot, 10},
	"tools":                        {kindRoot, 10},
}

func TestEveryPackageIsClassified(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, file := range productionFiles(t) {
		seen[file.pkgDir] = true
		if _, ok := packages[file.pkgDir]; !ok {
			t.Errorf("%s: package %s is not classified in packages", file.path, file.pkgDir)
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
				t.Errorf("%s (layer %d) imports %s (layer %d): imports must point to a lower layer",
					file.path, packages[file.pkgDir].layer, imported, packages[imported].layer)
			}
		}
	}
}

func TestFoundationsImportOnlyFoundations(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		if packages[file.pkgDir].kind != kindFoundation {
			continue
		}
		for _, imported := range modulePackages(file) {
			if packages[imported].kind != kindFoundation {
				t.Errorf("%s: foundation imports %s (%s)", file.path, imported, packages[imported].kind)
			}
		}
	}
}

func TestSurfacesAreImportedOnlyBySurfaces(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		kind := packages[file.pkgDir].kind
		if kind == kindSurface || kind == kindRoot {
			continue
		}
		for _, imported := range modulePackages(file) {
			if packages[imported].kind == kindSurface {
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

// layerViolation reports an import that does not point to a strictly lower layer.
func layerViolation(owner, imported string) bool {
	return packages[owner].layer <= packages[imported].layer
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
		if used[file.pkgDir] == nil {
			used[file.pkgDir] = map[string]bool{}
		}
		for _, imported := range modulePackages(file) {
			used[file.pkgDir][strings.TrimPrefix(imported, "internal/")] = true
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
