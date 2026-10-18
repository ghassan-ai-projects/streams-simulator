package architecture

import (
	"go/ast"
	"path"
	"slices"
	"strings"
	"testing"
)

// The relay passes a hardware gateway's records up and a host's commands down untouched
// (the physical devices design, relay mode): no perturbation, wire fault or fault injection acts
// on them, and no option can add one. The relay's own files may therefore name none of that code.

const plantwireDir = "internal/plantwire"

var (
	relayEdgeFiles       = []string{"gateways.go", "ticks.go"}
	relayForbiddenModule = []string{"internal/perturb", "internal/world", "internal/truth", "internal/score", "internal/audit", "internal/suite"}
	relayForbiddenName   = []string{"WireFault", "ApplyWireFaults", "FaultDrop", "FaultDuplicate", "FaultSwap", "FaultGarble", "Perturb", "Perturbation"}
)

func isRelayFile(file productionFile) bool {
	module, layer := splitModule(file.pkgDir)
	if module != plantwireDir {
		return false
	}
	base := path.Base(file.path)
	if layer == "uds" {
		return slices.Contains(relayEdgeFiles, base)
	}
	return (layer == "domain" || layer == "app") && strings.HasPrefix(base, "relay")
}

func relayPassThroughViolations(files []productionFile) []string {
	var violations []string
	for _, file := range files {
		if !isRelayFile(file) {
			continue
		}
		for _, imported := range modulePackages(file) {
			if module, _ := splitModule(imported); slices.Contains(relayForbiddenModule, module) {
				violations = append(violations, file.path+" imports "+imported+": the relay changes nothing it passes on")
			}
		}
		violations = append(violations, relayForbiddenNames(file)...)
	}
	return violations
}

func relayForbiddenNames(file productionFile) []string {
	var violations []string
	ast.Inspect(file.source, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok && slices.Contains(relayForbiddenName, ident.Name) {
			violations = append(violations, file.path+" names "+ident.Name+": no wire fault or perturbation reaches a relayed record")
		}
		return true
	})
	return violations
}

func TestTheRelayImportsNoPerturbationOrFaultCode(t *testing.T) {
	t.Parallel()
	files := productionFiles(t)
	relayFiles := 0
	for _, file := range files {
		if isRelayFile(file) {
			relayFiles++
		}
	}
	if relayFiles < 8 {
		t.Fatalf("found %d relay files, want the relay's domain, app and edge files", relayFiles)
	}
	for _, violation := range relayPassThroughViolations(files) {
		t.Error(violation)
	}
}

func TestTheRelayGuardFlagsPerturbationAndWireFaultCode(t *testing.T) {
	t.Parallel()
	cases := map[string]struct{ dir, name, source, want string }{
		"a perturbation import": {plantwireDir + "/internal/app", "relay_x.go",
			"package app\n\nimport _ \"" + mod + "internal/perturb\"\n", "imports internal/perturb"},
		"a wire fault": {plantwireDir + "/internal/domain", "relay_x.go",
			"package domain\n\nvar _ = ApplyWireFaults\n", "names ApplyWireFaults"},
		"an edge file": {plantwireDir + "/internal/uds", "gateways.go",
			"package uds\n\nvar _ = WireFault{}\n", "names WireFault"},
	}
	for name, tc := range cases {
		violations := relayPassThroughViolations([]productionFile{cliSnippet(t, tc.dir, tc.name, tc.source)})
		if len(violations) != 1 || !strings.Contains(violations[0], tc.want) {
			t.Errorf("%s: violations = %v, want one containing %q", name, violations, tc.want)
		}
	}
	node := cliSnippet(t, plantwireDir+"/internal/app", "session.go", "package app\n\nvar _ = WireFault{}\n")
	if violations := relayPassThroughViolations([]productionFile{node}); len(violations) != 0 {
		t.Errorf("the node session was flagged: %v", violations)
	}
}
