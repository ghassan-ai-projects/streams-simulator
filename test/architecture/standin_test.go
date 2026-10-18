package architecture

import "testing"

// standinPackage plays a scripted plant wire node for tests. It is production-shaped
// code only so that tests of several packages of the plantwire module can share it.
const standinPackage = "internal/plantwire/internal/standin"

func TestTheStandInNodeIsImportedByTestsOnly(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		if file.pkgDir == standinPackage {
			continue
		}
		for _, imported := range modulePackages(file) {
			if imported == standinPackage {
				t.Errorf("%s imports %s: a scripted node is test support and must never ship in a package that could be taken for hardware", file.path, imported)
			}
		}
	}
}
