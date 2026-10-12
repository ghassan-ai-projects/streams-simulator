package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePrefix = "github.com/ghassan-ai-projects/streams-simulator/"

// Direct dependencies reflect ownership, not transitive reachability. Changing
// this map requires reviewing the design and the importing package's role.
var packageDependencies = map[string]string{
	".":                           "",
	"tools":                       "",
	"cmd/streamsim":               "cli",
	"internal/cli":                "adapter device deviceworld domain grading mcp model plantwire refconsumer release report run suite",
	"internal/director":           "adapter audit domain model run score truth world",
	"internal/mcp":                "director model schemas",
	"internal/run":                "adapter canonical domain journal model perturb sink wall world",
	"internal/grading":            "domain model run score world",
	"internal/report":             "grading journal model run",
	"internal/release":            "canonical",
	"internal/score":              "domain model world",
	"internal/suite":              "audit domain model perturb randutil truth world",
	"internal/audit":              "domain model perturb world",
	"internal/refconsumer":        "canonical model",
	"internal/deviceworld":        "device domain model world",
	"internal/device":             "canonical jsonschema device/contract",
	"internal/plantwire":          "canonical domain jsonschema model plantwire/contract run world",
	"internal/plantwire/contract": "",
	"internal/world":              "domain jsonschema model randutil",
	"internal/truth":              "domain model world",
	"internal/perturb":            "domain model randutil",
	"internal/adapter":            "canonical jsonschema model schemas",
	"internal/domain":             "canonical jsonschema model schemas",
	"internal/model":              "jsonschema schemas",
	"internal/journal":            "model",
	"internal/jsonschema":         "canonical",
	"internal/canonical":          "",
	"internal/device/contract":    "",
	"internal/randutil":           "",
	"internal/schemas":            "",
	"internal/sink":               "",
	"internal/wall":               "",
	"internal/testsupport":        "",
}

func TestPackageDependencies(t *testing.T) {
	t.Parallel()
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect dependency in %s: %w", path, err)
		}
		if entry.IsDir() && isAgentWorkspace(path) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("inspect dependency in %s: %w", path, err)
		}
		owner, err := filepath.Rel("../..", filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("inspect dependency in %s: %w", path, err)
		}
		owner = filepath.ToSlash(owner)
		if owner == "." && filepath.Base(path) == "tools.go" {
			owner = "tools" // Build-tagged development tools, outside runtime.
		}
		owner, _ = splitModule(owner)
		if _, exists := packageDependencies[owner]; !exists {
			t.Errorf("unreviewed package: %s", owner)
		}
		for _, spec := range source.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("inspect dependency in %s: %w", path, err)
			}
			if dependencyAllowed(owner, imported) {
				continue
			}
			t.Errorf("%s imports %s outside its responsibility; review BAR.md before adding an edge", path, imported)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func dependencyAllowed(owner, imported string) bool {
	if !strings.Contains(strings.Split(imported, "/")[0], ".") {
		return true
	}
	if owner == "tools" {
		return imported == "golang.org/x/tools/cmd/deadcode" ||
			imported == "golang.org/x/tools/cmd/goimports" ||
			imported == "golang.org/x/vuln/cmd/govulncheck"
	}
	if strings.HasPrefix(imported, modulePrefix) {
		allowed, exists := packageDependencies[owner]
		if !exists {
			return false
		}
		dir := strings.TrimPrefix(imported, modulePrefix)
		if module, layer := splitModule(dir); module == owner && layer != "" {
			return true // a module's own internal layers
		}
		module, _ := splitModule(dir)
		target := strings.TrimPrefix(module, "internal/")
		for _, name := range strings.Fields(allowed) {
			if name == target {
				return true
			}
		}
		return false
	}
	// Only application surfaces speak MCP. Foundations and the deterministic
	// pipeline do not gain external dependencies through transport convenience.
	return imported == "github.com/modelcontextprotocol/go-sdk/mcp" &&
		(owner == "internal/mcp" || owner == "internal/refconsumer")
}

func TestDependencyGuardRejectsUpwardAndExternalEdges(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		owner, imported string
		allowed         bool
	}{
		{"internal/world", modulePrefix + "internal/model", true},
		{"internal/world", modulePrefix + "internal/run", false},
		{"internal/model", modulePrefix + "internal/device", false},
		{"internal/sink", modulePrefix + "internal/world", false},
		{"internal/world", "github.com/modelcontextprotocol/go-sdk/mcp", false},
		{"internal/mcp", "github.com/modelcontextprotocol/go-sdk/mcp", true},
		{"internal/domain", "example.com/new-dependency", false},
		{"internal/unreviewed", modulePrefix + "internal/model", false},
		{".", modulePrefix + "internal/mcp", false},
		{"tools", "golang.org/x/tools/cmd/deadcode", true},
		{"tools", "github.com/modelcontextprotocol/go-sdk/mcp", false},
	} {
		t.Run(tc.owner+"/"+tc.imported, func(t *testing.T) {
			t.Parallel()
			if got := dependencyAllowed(tc.owner, tc.imported); got != tc.allowed {
				t.Fatalf("allowed=%v, want %v", got, tc.allowed)
			}
		})
	}
}
