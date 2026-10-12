package architecture

import (
	"slices"
	"strings"
	"testing"
)

// The MCP surface holds protocol only: tool definitions and schemas, argument
// decoding from tool calls, the mapping of refusals to the wire and the
// transport that carries them. Every decision about a world, a run, truth or
// scoring lives in the director module, and mcp reaches it only through the
// director's facade. The rule is the package list below plus one import rule:
// beyond the director, mcp imports foundations and nothing else.

const mcpDir = "internal/mcp"

// mcpLayers are the only packages the MCP module may contain besides its
// facade.
var mcpLayers = []string{"protocol", "transport"}

func TestMCPHoldsOnlyProtocol(t *testing.T) {
	t.Parallel()
	var files []productionFile
	for _, file := range productionFiles(t) {
		if module, _ := splitModule(file.pkgDir); module == mcpDir {
			files = append(files, file)
		}
	}
	for _, violation := range mcpViolations(files) {
		t.Error(violation)
	}
}

// mcpViolations returns every way the MCP module's files exceed protocol: a
// package that is neither the facade nor protocol or transport, and an import
// of any module package that is neither the director nor a foundation.
func mcpViolations(files []productionFile) []string {
	var violations []string
	for _, file := range files {
		if _, layer := splitModule(file.pkgDir); layer != "" && !slices.Contains(mcpLayers, layer) {
			violations = append(violations, file.path+": mcp has no package "+file.pkgDir+"; put the logic in the director and call its facade")
		}
		for _, imported := range modulePackages(file) {
			if !mcpMayImport(imported) {
				violations = append(violations, file.path+": imports "+imported+"; mcp decodes and frames, the director decides, so reach it through internal/director")
			}
		}
	}
	return violations
}

func mcpMayImport(imported string) bool {
	module, _ := splitModule(imported)
	if module == mcpDir || module == "internal/director" {
		return true
	}
	return infoOf(imported).kind == kindFoundation
}

func TestMCPRulesRejectBusinessLogic(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		dir, source, want string
	}{
		"a use case in the facade": {mcpDir, `package mcp

import "` + mod + `internal/run"

func Open() { _ = run.MaxRunSpanNS }
`, "imports internal/run"},
		"scoring in a tool handler": {mcpDir + "/internal/protocol", `package protocol

import "` + mod + `internal/score"

func handle() { _ = score.Score }
`, "imports internal/score"},
		"truth in the transport": {mcpDir + "/internal/transport", `package transport

import "` + mod + `internal/truth"

func serve() { _ = truth.NewStore }
`, "imports internal/truth"},
		"a world import": {mcpDir + "/internal/protocol", `package protocol

import "` + mod + `internal/world"

func handle() { _ = world.New }
`, "imports internal/world"},
		"an audit import": {mcpDir + "/internal/protocol", `package protocol

import "` + mod + `internal/audit"

func handle() { _ = audit.NewPanel }
`, "imports internal/audit"},
		"an application layer": {mcpDir + "/internal/app", `package app

func Create() {}
`, "mcp has no package " + mcpDir + "/internal/app"},
		"an entropy edge": {mcpDir + "/internal/capability", `package capability

func NewToken() {}
`, "mcp has no package " + mcpDir + "/internal/capability"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			violations := mcpViolations([]productionFile{cliSnippet(t, tc.dir, "injected.go", tc.source)})
			if !slices.ContainsFunc(violations, func(v string) bool { return strings.Contains(v, tc.want) }) {
				t.Fatalf("violations = %v, want one containing %q", violations, tc.want)
			}
		})
	}
}

func TestMCPRulesAcceptProtocol(t *testing.T) {
	t.Parallel()
	source := `package protocol

import (
	"` + mod + `internal/director"
	"` + mod + `internal/model"
)

func handle(d *director.Director) { _ = model.SimVersion }
`
	if violations := mcpViolations([]productionFile{cliSnippet(t, mcpDir+"/internal/protocol", "handle.go", source)}); len(violations) != 0 {
		t.Fatalf("violations = %v", violations)
	}
}
