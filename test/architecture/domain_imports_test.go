package architecture

import (
	"slices"
	"strings"
	"testing"
)

// domainLayerEdgeImports lists the reviewed imports of an edge-core module
// from a domain layer: value types or pure parsing only, never an operation
// backed by an edge.
var domainLayerEdgeImports = map[string][]string{
	"internal/grading/internal/domain":     {"internal/run"},
	"internal/report/internal/domain":      {"internal/journal"},
	"internal/deviceworld/internal/domain": {"internal/device"},
}

const modulePath = "github.com/ghassan-ai-projects/streams-simulator/"

func TestDomainLayersImportOnlyPureModules(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		if _, layer := splitModule(file.pkgDir); layer != "domain" {
			continue
		}
		for _, imported := range file.paths {
			if reason := impureImport(file.pkgDir, imported); reason != "" {
				t.Errorf("%s imports %s: %s", file.path, imported, reason)
			}
		}
	}
}

func impureImport(pkgDir, imported string) string {
	target, internal := strings.CutPrefix(imported, modulePath)
	if !internal || strings.HasPrefix(target, pkgDir[:strings.Index(pkgDir, "/internal/domain")]+"/") {
		return ""
	}
	info := packages[target]
	switch {
	case info.kind == kindFoundation, info.kind == kindCore, target == "internal/domain":
		return ""
	case slices.Contains(domainLayerEdgeImports[pkgDir], target):
		return ""
	}
	return "a domain layer imports only foundations, pure cores and reviewed value types (domainLayerEdgeImports)"
}
