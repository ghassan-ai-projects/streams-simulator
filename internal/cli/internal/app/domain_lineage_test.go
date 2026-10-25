package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func lineageWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for source, rel := range map[string]string{
		testsupport.FactoryDomain():      testsupport.WorkspaceDomain("factory-utilities"),
		testsupport.FactoryComposition(): "compositions/factory-utilities.composition.json",
		testsupport.TwinThermalChoices(): "twins/thermal-board.twin.json",
	} {
		raw, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, rel)), 0o750); err != nil {
			t.Fatal(err)
		}
		testsupport.WriteFileIn(t, filepath.Join(root, filepath.Dir(rel)), filepath.Base(rel), raw)
	}
	return root
}

func lineageArguments(root string, extra ...string) []string {
	return append([]string{"domain", "lineage", "--domains-dir", testsupport.WorkspaceFolder(root, "domains"), "--compositions-dir", filepath.Join(root, "compositions"),
		"--twins-dir", filepath.Join(root, "twins")}, extra...)
}

func TestDomainLineageNamesTheCompositionThatProducesAPlantAndIsTheSameEveryTime(t *testing.T) {
	t.Parallel()
	root := lineageWorkspace(t)
	first := invoke(t, lineageArguments(root)...).mustSucceed(t)
	doc := first.document(t, "domain_lineage")
	plants, _ := doc["plants"].([]any)
	plant, _ := plants[0].(map[string]any)
	composition, _ := plant["composition"].(map[string]any)
	if len(plants) != 1 || composition["status"] != "in_sync" {
		t.Fatalf("lineage = %v, want the factory composition in sync", doc)
	}
	if second := invoke(t, lineageArguments(root)...).mustSucceed(t); second.stdout != first.stdout {
		t.Fatal("two runs printed different documents")
	}
}

func TestDomainLineageOfOneDomainRefusesAnIdItDoesNotHold(t *testing.T) {
	t.Parallel()
	root := lineageWorkspace(t)
	invoke(t, lineageArguments(root, "--domain", "factory-utilities")...).mustSucceed(t).document(t, "domain_lineage")
	if failure := invoke(t, lineageArguments(root, "--domain", "no-such-plant")...).failure(t); failure["code"] != "domain_not_found" {
		t.Fatalf("failure = %v, want domain_not_found", failure)
	}
}

func TestDomainLineageRefusesAFolderThatIsMissingAndArgumentsItDoesNotTake(t *testing.T) {
	t.Parallel()
	root := lineageWorkspace(t)
	for name, tc := range map[string]struct {
		args []string
		code string
	}{
		"no domains folder":  {[]string{"domain", "lineage"}, "usage_invalid"},
		"a missing folder":   {lineageArguments(root, "--runs-dir", filepath.Join(root, "runs")), "file_not_found"},
		"a positional value": {lineageArguments(root, "extra"), "usage_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if failure := invoke(t, tc.args...).failure(t); failure["code"] != tc.code {
				t.Fatalf("failure = %v, want code %s", failure, tc.code)
			}
		})
	}
}

func TestDomainLineageRefusesADomainFileThatBreaksItsContract(t *testing.T) {
	t.Parallel()
	root := lineageWorkspace(t)
	testsupport.WriteFileIn(t, testsupport.WorkspaceFolder(root, "domains"), "bad.domain.json", []byte(`{"id":"X"}`))
	if failure := invoke(t, lineageArguments(root)...).failure(t); failure["code"] != "domain_invalid" {
		t.Fatalf("failure = %v, want domain_invalid", failure)
	}
}
