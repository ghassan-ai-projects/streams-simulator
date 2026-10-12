package architecture

import (
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// cliSnippet parses source as a file of the CLI package at dir.
func cliSnippet(t *testing.T, dir, name, source string) productionFile {
	t.Helper()
	fset := token.NewFileSet()
	file := dir + "/" + name
	parsed, err := parser.ParseFile(fset, file, source, parser.ParseComments)
	if err != nil {
		t.Fatalf("snippet does not parse: %v", err)
	}
	return productionFile{path: file, pkgDir: dir, fset: fset, source: parsed, paths: importPaths(parsed), imports: importNames(parsed)}
}

func TestCLIWiringRulesRejectBusinessLogic(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		dir, file, source string
		want              string
	}{
		"a new file of logic": {cliAppDir, "ledger.go", `package app

func loadLedger(path string) {}
`, "not a wiring file"},
		"a nested package under the command layer": {cliAppDir + "/logic", "logic.go", `package logic

func Decode() {}
`, "the CLI has no package " + cliAppDir + "/logic"},
		"a package the CLI never had": {cliDir + "/internal/serve", "serve.go", `package serve

func Serve() {}
`, "the CLI has no package " + cliDir + "/internal/serve"},
		"decoding in a command": {cliAppDir, "catalog.go", `package app

import "encoding/json"

func decode(raw []byte) error { return json.Unmarshal(raw, &struct{}{}) }
`, "imports encoding/json"},
		"signing in a command": {cliAppDir, "manifest.go", `package app

import "crypto/ed25519"

func sign(seed []byte) []byte { return ed25519.NewKeyFromSeed(seed) }
`, "imports crypto/ed25519"},
		"file access in a command": {cliAppDir, "score.go", `package app

import "os"

func read(path string) ([]byte, error) { return os.ReadFile(path) }
`, "imports os"},
		"a module's layer": {cliAppDir, "run.go", `package app

import "` + mod + `internal/run/internal/app"

var _ = app.Config{}
`, "imports " + mod + "internal/run/internal/app"},
		"a facade another file wires": {cliAppDir, "catalog.go", `package app

import "` + mod + `internal/grading"

var _ = grading.Grade
`, "imports " + mod + "internal/grading"},
		"a facade the CLI does not wire": {cliAppDir, "score.go", `package app

import "` + mod + `internal/truth"

var _ = truth.NewStore
`, "imports " + mod + "internal/truth"},
		"serving": {cliAppDir, "mcp.go", `package app

func serve() { go func() {}() }
`, "concurrency and resource handling"},
		"an exported rule": {cliAppDir, "score.go", `package app

func GradeRun() {}
`, "exports GradeRun"},
		"logic appended to a wiring file": {cliAppDir, "score.go", "package app\n" + strings.Repeat("\n", 80), "exceeds its line cap"},
		"a module in a CLI edge": {cliF, "files.go", `package files

import "` + mod + `internal/run"

var _ = run.New
`, "imports " + mod + "internal/run"},
		"a socket in a CLI edge": {cliP, "process.go", `package process

import "net/http"

var _ = http.StatusOK
`, "imports net/http"},
		"a CLI edge that grows": {cliF, "files.go", "package files\n" + strings.Repeat("\n", 50), "has more lines than its cap"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			violations := cliViolations([]productionFile{cliSnippet(t, tc.dir, tc.file, tc.source)})
			if !slices.ContainsFunc(violations, func(v string) bool { return strings.Contains(v, tc.want) }) {
				t.Fatalf("violations = %v, want one containing %q", violations, tc.want)
			}
		})
	}
}

func TestCLIWiringRulesAcceptWiring(t *testing.T) {
	t.Parallel()
	source := `package app

import (
	"flag"
	"fmt"

	"` + mod + `internal/run"
)

func cmdWiring(args []string) (any, error) {
	fs := flag.NewFlagSet("wiring", flag.ContinueOnError)
	return fmt.Sprint(fs.Name(), run.MaxRunSpanNS), nil
}
`
	if violations := cliViolations([]productionFile{cliSnippet(t, cliAppDir, "replay.go", source)}); len(violations) != 0 {
		t.Fatalf("violations = %v", violations)
	}
}
