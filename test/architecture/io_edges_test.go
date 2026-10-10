package architecture

import (
	"go/ast"
	"slices"
	"strings"
	"testing"
)

// ioImports are the imports that give a file access to the outside world.
var ioImports = []string{
	"os", "os/exec", "os/signal", "net", "net/http",
	"crypto/rand", "math/rand", "math/rand/v2", "syscall", "io/ioutil",
}

// clockCalls are the time functions that read or wait on the wall clock.
var clockCalls = []string{
	"Now", "Since", "Until", "Sleep", "After", "AfterFunc", "NewTimer", "NewTicker", "Tick",
}

// ioEdge declares why a production file may touch files, sockets, processes,
// entropy or the wall clock. debt names the round that moves the use to an
// edge package; foundation and core packages may only hold debt.
type ioEdge struct {
	uses []string
	why  string
	debt string
}

// ioEdges is the complete inventory of I/O and clock sites (STANDARD M3, M4).
// A file that uses a capability not listed here fails the gate, and a listed
// capability a file no longer uses fails it too, so the table only shrinks.
var ioEdges = map[string]ioEdge{
	"cmd/streamsim/main.go":            {uses: []string{"import:os"}, why: "process exit"},
	"cmd/streamsim/function_length.go": {uses: []string{"call:filepath.WalkDir", "import:os"}, why: "development tool, build-ignored"},
	"internal/wall/wall_clock.go":      {uses: []string{"call:time.Now"}, why: "the declared wall-clock seam"},
	"internal/sink/sink.go":            {uses: []string{"import:net/http", "import:os"}, why: "file and http-push sinks"},
	"internal/mcp/world.go":            {uses: []string{"import:crypto/rand"}, why: "operator token entropy"},
	"internal/cli/arguments.go":        {uses: []string{"call:time.Now"}, why: "scripted command id (deferred D-19)"},
	"internal/cli/catalog.go":          {uses: []string{"import:os"}, why: "catalog and adapter directory scans"},
	"internal/cli/cli.go":              {uses: []string{"import:os"}, why: "stdout, stderr and exit"},
	"internal/cli/device.go":           {uses: []string{"import:net", "import:os", "import:os/signal", "import:syscall"}, why: "device serve process lifecycle"},
	"internal/cli/device_world.go":     {uses: []string{"call:time.Now", "call:time.Since", "import:os"}, why: "device-world wiring"},
	"internal/cli/manifest.go":         {uses: []string{"import:os"}, why: "release manifest files"},
	"internal/cli/manifest_options.go": {uses: []string{"call:time.Now"}, why: "manifest creation time"},
	"internal/cli/mcp.go":              {uses: []string{"go-statement", "import:net", "import:net/http", "import:os"}, why: "MCP stdio and operator listener"},
	"internal/cli/refconsumer.go":      {uses: []string{"import:os"}, why: "trace and verdict files"},
	"internal/cli/score.go":            {uses: []string{"import:os"}, why: "artifact files"},
	"internal/cli/suite.go":            {uses: []string{"import:os"}, why: "suite output files"},

	"internal/domain/file.go":                {uses: []string{"import:os"}, why: "domain file and directory loading; the only file-system edge of the package"},
	"internal/adapter/adapter.go":            {uses: []string{"import:os"}, why: "adapter file loading", debt: "R4"},
	"internal/adapter/fixture.go":            {uses: []string{"import:os"}, why: "conformance fixture file", debt: "R4"},
	"internal/adapter/verify_conformance.go": {uses: []string{"import:os"}, why: "conformance golden and schema files", debt: "R4"},
	"internal/device/uds.go":                 {uses: []string{"go-statement", "import:net", "import:os"}, why: "unix socket listener and session", debt: "R9"},
	"internal/run/artifact.go":               {uses: []string{"import:os"}, why: "artifact files", debt: "R11"},
	"internal/run/artifact_metadata.go":      {uses: []string{"call:time.Now"}, why: "artifact created_at", debt: "R11"},
	"internal/run/finalize.go":               {uses: []string{"import:os"}, why: "artifact publication", debt: "R11"},
	"internal/run/initialize.go":             {uses: []string{"import:os"}, why: "durable ledger file", debt: "R11"},
	"internal/run/replay.go":                 {uses: []string{"import:os"}, why: "artifact loading", debt: "R11"},
	"internal/run/run.go":                    {uses: []string{"call:time.NewTimer", "import:os"}, why: "quiescence timer and ledger handle", debt: "R11"},
	"internal/run/verdict.go":                {uses: []string{"call:time.Now"}, why: "unblinded timestamp", debt: "R11"},
}

func TestIOStaysInDeclaredEdges(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, file := range productionFiles(t) {
		used := ioUses(file)
		seen[file.path] = true
		declared := ioEdges[file.path].uses
		if !slices.Equal(used, declared) {
			t.Errorf("%s: I/O uses %v, declared %v: move the code to an edge or update ioEdges with a reason",
				file.path, used, declared)
		}
	}
	for path := range ioEdges {
		if !seen[path] {
			t.Errorf("stale ioEdges entry: %s is not a production file", path)
		}
	}
}

func TestPureKindsHoldOnlyScheduledIODebt(t *testing.T) {
	t.Parallel()
	for path, edge := range ioEdges {
		kind := packages[packageDirectory(path)].kind
		pure := kind == kindFoundation || kind == kindCore
		if pure && edge.debt == "" {
			t.Errorf("%s: %s package declares I/O without scheduled debt", path, kind)
		}
		if edge.debt != "" && !roundIsPlanned(t, edge.debt) {
			t.Errorf("%s: debt %q must name a round row in PLAN.md", path, edge.debt)
		}
		if edge.why == "" {
			t.Errorf("%s: ioEdges entry needs a reason", path)
		}
	}
}

// fsCalls are path/filepath functions that touch the file system or process
// working directory; contextClockCalls are context constructors that read the
// wall clock.
var (
	fsCalls           = []string{"Abs", "EvalSymlinks", "Glob", "Walk", "WalkDir"}
	contextClockCalls = []string{"WithTimeout", "WithDeadline"}
)

// ioUses lists the sorted I/O, clock and concurrency capabilities a file uses.
// Function values such as `Clock: time.Now` count as uses.
func ioUses(file productionFile) []string {
	uses := map[string]bool{}
	for _, path := range file.paths {
		if slices.Contains(ioImports, path) {
			uses["import:"+path] = true
		}
	}
	if file.imports["."] == "time" {
		uses["import:time(dot)"] = true
	}
	ast.Inspect(file.source, func(node ast.Node) bool {
		if use, ok := capabilityOf(file, node); ok {
			uses[use] = true
		}
		return true
	})
	list := make([]string, 0, len(uses))
	for use := range uses {
		list = append(list, use)
	}
	slices.Sort(list)
	if len(list) == 0 {
		return nil
	}
	return list
}

func capabilityOf(file productionFile, node ast.Node) (string, bool) {
	switch n := node.(type) {
	case *ast.GoStmt:
		return "go-statement", true
	case *ast.SelectorExpr:
		return selectorCapability(file, n)
	}
	return "", false
}

func selectorCapability(file productionFile, selector *ast.SelectorExpr) (string, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	name := selector.Sel.Name
	switch file.imports[ident.Name] {
	case "time":
		return "call:time." + name, slices.Contains(clockCalls, name)
	case "path/filepath":
		return "call:filepath." + name, slices.Contains(fsCalls, name)
	case "context":
		return "call:context." + name, slices.Contains(contextClockCalls, name)
	}
	return "", false
}

// roundIsPlanned reports whether PLAN.md has a status row for the round.
func roundIsPlanned(t *testing.T, round string) bool {
	t.Helper()
	plan, err := repositoryRoot(t).ReadFile("docs/refactoring/modularity-20261010/PLAN.md")
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(plan), "| "+round+" |")
}
