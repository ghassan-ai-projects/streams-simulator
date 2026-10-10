package architecture

import (
	"slices"
	"strings"
	"testing"
)

// packageMapDocument lists every package with its responsibility (STANDARD M1).
const packageMapDocument = ".agents/context/architecture.md"

func TestEveryPackageDocumentsItsResponsibility(t *testing.T) {
	t.Parallel()
	documented := map[string]bool{}
	for _, file := range productionFiles(t) {
		if statesResponsibility(file) {
			documented[file.pkgDir] = true
		}
	}
	for _, file := range productionFiles(t) {
		if !documented[file.pkgDir] {
			t.Errorf("package %s has no package comment stating its responsibility", file.pkgDir)
			documented[file.pkgDir] = true // report each package once
		}
	}
}

func TestPackageMapListsEveryPackage(t *testing.T) {
	t.Parallel()
	document, err := repositoryRoot(t).ReadFile(packageMapDocument)
	if err != nil {
		t.Fatal(err)
	}
	text := string(document)
	var directories []string
	for _, file := range productionFiles(t) {
		directories = append(directories, file.pkgDir)
	}
	slices.Sort(directories)
	for _, directory := range slices.Compact(directories) {
		if directory == "." || directory == "tools" {
			continue
		}
		if !strings.Contains(text, "`"+directory+"`") && !strings.Contains(text, "`"+directory+"/") {
			t.Errorf("%s does not list package %s", packageMapDocument, directory)
		}
	}
}

// statesResponsibility reports a package comment that begins "Package <name>"
// (or "Command <name>" for a main package), not a license header.
func statesResponsibility(file productionFile) bool {
	if file.source.Doc == nil {
		return false
	}
	text := strings.TrimSpace(file.source.Doc.Text())
	name := file.source.Name.Name
	return strings.HasPrefix(text, "Package "+name) || strings.HasPrefix(text, "Command ")
}
