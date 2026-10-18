package architecture

import "testing"

// standinPackage plays a scripted plant wire node for tests. It is production-shaped
// code only so that tests of several packages of the plantwire module can share it.
const standinPackage = "internal/plantwire/internal/standin"

func standinImportViolations(files []productionFile) []string {
	var violations []string
	for _, file := range files {
		if file.pkgDir == standinPackage {
			continue
		}
		for _, imported := range modulePackages(file) {
			if imported == standinPackage {
				violations = append(violations, file.path+" imports "+imported+": a scripted node is test support and must never ship in a package that could be taken for hardware")
			}
		}
	}
	return violations
}

func TestTheStandInNodeIsImportedByTestsOnly(t *testing.T) {
	t.Parallel()
	for _, violation := range standinImportViolations(productionFiles(t)) {
		t.Error(violation)
	}
}

func TestTheStandInGuardFlagsAProductionFileThatImportsTheStandIn(t *testing.T) {
	t.Parallel()
	source := `package app

import "` + mod + standinPackage + `"

var _ = standin.New
`
	cases := map[string]string{"the plantwire app layer": "internal/plantwire/internal/app", "the cli": "internal/cli/internal/app", "the facade": "internal/plantwire"}
	for name, dir := range cases {
		violations := standinImportViolations([]productionFile{cliSnippet(t, dir, "x.go", source)})
		if len(violations) != 1 {
			t.Errorf("%s: violations = %v, want the import flagged", name, violations)
		}
	}
	own := cliSnippet(t, standinPackage, "own.go", "package standin\n")
	if violations := standinImportViolations([]productionFile{own}); len(violations) != 0 {
		t.Errorf("the stand-in's own files were flagged: %v", violations)
	}
}
