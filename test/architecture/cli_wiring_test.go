package architecture

import (
	"go/ast"
	"path"
	"slices"
	"strings"
	"testing"
)

// The command-line surface holds wiring only: flag parsing, turning flags into
// a scenario, the command registry, the JSON documents it prints and the
// failure documents it classifies. Everything that decides, loads, signs,
// replays, scores or serves lives in the module that owns it, and the CLI calls
// that module's facade. A rule this concrete cannot be argued around: a package
// that is not on the list, a file that is not on the list, an import the file
// has no use for, a construct that serves, or growth past the file's cap fails
// the build.

const (
	cliDir    = "internal/cli"
	cliAppDir = "internal/cli/internal/app"
)

const (
	mod  = modulePrefix // shortens the module paths below
	cliF = "internal/cli/internal/files"
	cliP = "internal/cli/internal/process"
)

// cliFile declares what one command-layer file is for: its wiring role, the
// module facades (and CLI edges) it may import, and its line cap, which is its
// current size plus a little headroom. A new file, a new import or growth past
// the cap is a reviewed change here.
type cliFile struct {
	role    string
	imports []string
	lines   int
}

var cliAppFiles = map[string]cliFile{
	"cli.go":               {"command registry and JSON output", []string{cliP}, 190},
	"arguments.go":         {"option parsing", nil, 190},
	"device_options.go":    {"option parsing", nil, 55},
	"manifest_options.go":  {"option parsing", nil, 45},
	"run_options.go":       {"option parsing", []string{"internal/model", "internal/run"}, 140},
	"run_script.go":        {"scenario from flags", []string{"internal/model", "internal/run"}, 190},
	"failure_codes.go":     {"failure document", []string{"internal/adapter", "internal/domain", "internal/model", "internal/run"}, 105},
	"failure_document.go":  {"failure document", []string{"internal/model"}, 80},
	"failures.go":          {"failure document", nil, 80},
	"run_flag_issues.go":   {"failure document", []string{"internal/model"}, 65},
	"catalog.go":           {"command: catalog, domain and adapter", []string{"internal/adapter", "internal/domain", "internal/model"}, 220},
	"domain_lint.go":       {"command: domain lint", []string{"internal/domain", "internal/lint"}, 75},
	"component.go":         {"command: component library and domain compose", []string{cliF, "internal/component", "internal/domain"}, 150},
	"device.go":            {"command: device serve", []string{"internal/device", "internal/deviceworld"}, 100},
	"manifest.go":          {"command: manifest", []string{"internal/release"}, 40},
	"mcp.go":               {"command: mcp", []string{"internal/mcp"}, 60},
	"refconsumer.go":       {"command: refconsumer", []string{"internal/refconsumer"}, 80},
	"replay.go":            {"command: replay and verify", []string{"internal/run"}, 85},
	"run.go":               {"command: run", []string{cliF, "internal/model", "internal/run"}, 95},
	"scenario.go":          {"command: scenario from-run", []string{cliF, "internal/run"}, 85},
	"report.go":            {"command: report", []string{cliF, "internal/report"}, 80},
	"score.go":             {"command: score", []string{"internal/grading"}, 70},
	"suite.go":             {"command: suite", []string{"internal/suite"}, 80},
	"plantwire.go":         {"command: plantwire node and host", []string{cliF, "internal/model", "internal/plantwire", "internal/run"}, 110},
	"plantwire_options.go": {"option parsing", []string{"internal/plantwire"}, 85},
}

// cliEdge declares what one of the CLI's own edges may import and how large it
// may grow. Edges carry the process and files; they know no simulator module.
type cliEdge struct {
	stdlib []string
	lines  int
}

var cliEdges = map[string]cliEdge{
	cliF: {[]string{"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"}, 40},
	cliP: {[]string{"io", "os", "time"}, 35},
}

// cliStdlib is the standard library the command layer needs: flags,
// formatting, errors, strings and paths. Decoding, hashing, signing, sockets,
// files and concurrency are not on it.
var cliStdlib = []string{
	"cmp", "context", "errors", "flag", "fmt", "io", "io/fs", "math", "path/filepath", "sort", "strconv", "strings", "time",
}

// cliJSONFiles are the files that print documents and so may marshal JSON.
var cliJSONFiles = []string{"cli.go"}

// cliExported is everything the command layer exports.
var cliExported = []string{"Build", "Main", "Run"}

func TestCLIHoldsOnlyWiring(t *testing.T) {
	t.Parallel()
	var files []productionFile
	for _, file := range productionFiles(t) {
		if file.pkgDir == cliDir || strings.HasPrefix(file.pkgDir, cliDir+"/") {
			files = append(files, file)
		}
	}
	for _, violation := range cliViolations(files) {
		t.Error(violation)
	}
}

// cliViolations returns every way the CLI's packages exceed wiring. The facade
// package is held to the facade rules elsewhere.
func cliViolations(files []productionFile) []string {
	var violations []string
	edgeLines := map[string]int{}
	for _, file := range files {
		switch edge, isEdge := cliEdges[file.pkgDir]; {
		case file.pkgDir == cliDir:
		case file.pkgDir == cliAppDir:
			violations = append(violations, cliAppFileViolations(file)...)
		case isEdge:
			violations = append(violations, cliEdgeFileViolations(file, edge)...)
			edgeLines[file.pkgDir] += cliLines(file)
		default:
			violations = append(violations, file.path+": the CLI has no package "+file.pkgDir+"; put the logic in the module that owns it and call its facade")
		}
	}
	for dir, lines := range edgeLines {
		if lines > cliEdges[dir].lines {
			violations = append(violations, dir+": has more lines than its cap; a CLI edge carries the process or files and nothing else")
		}
	}
	return violations
}

func cliLines(file productionFile) int { return file.fset.File(file.source.Pos()).LineCount() }

func cliAppFileViolations(file productionFile) []string {
	name := path.Base(file.path)
	entry, listed := cliAppFiles[name]
	if !listed {
		return []string{file.path + ": not a wiring file; put the logic in the module that owns it and call its facade, or list the file with its wiring role in cliAppFiles"}
	}
	var violations []string
	for _, imported := range file.paths {
		if !cliAppMayImport(imported, name, entry) {
			violations = append(violations, file.path+": imports "+imported+", which this file's wiring has no use for; call the owning module's facade from the file that wires it")
		}
	}
	if cliLines(file) > entry.lines {
		violations = append(violations, file.path+": exceeds its line cap; move the logic to the module that owns it")
	}
	return append(violations, cliConstructViolations(file)...)
}

func cliAppMayImport(imported, file string, entry cliFile) bool {
	if directory, internal := strings.CutPrefix(imported, modulePrefix); internal {
		return slices.Contains(entry.imports, directory)
	}
	if imported == "encoding/json" {
		return slices.Contains(cliJSONFiles, file)
	}
	return slices.Contains(cliStdlib, imported)
}

func cliEdgeFileViolations(file productionFile, edge cliEdge) []string {
	var violations []string
	for _, imported := range file.paths {
		if !slices.Contains(edge.stdlib, imported) {
			violations = append(violations, file.path+": imports "+imported+"; a CLI edge uses only "+strings.Join(edge.stdlib, ", ")+" and holds no simulator knowledge")
		}
	}
	return violations
}

// cliConstructViolations rejects what only a server or a rule engine needs:
// concurrency, deferred cleanup of resources the CLI must not hold, and an
// exported surface beyond the entry points.
func cliConstructViolations(file productionFile) []string {
	var violations []string
	ast.Inspect(file.source, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.GoStmt, *ast.SelectStmt, *ast.DeferStmt, *ast.ChanType:
			violations = append(violations, file.path+": "+file.fset.Position(n.Pos()).String()+": concurrency and resource handling belong to the module that owns the resource")
		case *ast.FuncDecl:
			if n.Recv == nil && n.Name.IsExported() && !slices.Contains(cliExported, n.Name.Name) {
				violations = append(violations, file.path+": exports "+n.Name.Name+"; the command layer exports only its entry points")
			}
		}
		return true
	})
	return violations
}
