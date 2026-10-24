package component

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type shippedWorkspace struct {
	root string
	dirs WorkspaceDirs
}

func copyTo(t *testing.T, source, root, rel string) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, filepath.Dir(target), filepath.Base(target), readFile(t, source))
}

func newShippedWorkspace(t *testing.T, library *Library) shippedWorkspace {
	t.Helper()
	root := t.TempDir()
	copyTo(t, testsupport.FactoryDomain(), root, testsupport.WorkspaceDomain("factory-utilities"))
	copyTo(t, testsupport.FactoryComposition(), root, "compositions/factory-utilities.composition.json")
	copyTo(t, testsupport.Domain("aquaculture-pond"), root, testsupport.WorkspaceDomain("aquaculture-pond"))
	copyTo(t, testsupport.TwinThermalChoices(), root, "twins/thermal-board.twin.json")
	for _, stem := range []string{"fan", "indicator", "thermal_sensor"} {
		copyTo(t, filepath.Join(testsupport.TwinThermalKindsDir(), stem+".kind.json"), root, "kinds/"+stem+".kind.json")
	}
	if err := os.MkdirAll(filepath.Join(root, "runs"), 0o750); err != nil {
		t.Fatal(err)
	}
	twin := thermalTwinInput(t, library)
	testsupport.WriteFileIn(t, testsupport.WorkspaceFolder(root, "domains"), "arduino-thermal-twin.domain.json", twin.Domain)
	return shippedWorkspace{root: root, dirs: WorkspaceDirs{Domains: testsupport.WorkspaceFolder(root, "domains"), Kinds: filepath.Join(root, "kinds"), Twins: filepath.Join(root, "twins"),
		Compositions: filepath.Join(root, "compositions"), Runs: filepath.Join(root, "runs")}}
}

func (w shippedWorkspace) addRun(t *testing.T, id, domain, digest string) {
	t.Helper()
	dir := filepath.Join(w.root, "runs", id)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, dir, "run.json", []byte(`{"mode":"explore","domain":{"id":"`+domain+`","version":"0.1.0","digest":"`+digest+`"}}`))
}

func TestLineageOfTheShippedWorkspaceMatchesItsGolden(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	workspace := newShippedWorkspace(t, library)
	current := "sha256:" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	workspace.addRun(t, "r-old", "factory-utilities", current)
	got, err := Lineage(library, workspace.dirs, "")
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, got, "lineage-shipped-workspace.golden.json")
}

func TestLineageReportsACompositionEditedAfterItsDomainWasComposed(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	workspace := newShippedWorkspace(t, library)
	before, err := Lineage(library, workspace.dirs, "factory-utilities")
	if err != nil {
		t.Fatal(err)
	}
	if composition := before.Plants[0].Composition; composition == nil || composition.Status != "in_sync" || before.Plants[0].Editable.Reason != "composition" {
		t.Fatalf("lineage = %+v, want the shipped composition in sync", before.Plants)
	}
	path := filepath.Join(workspace.dirs.Compositions, "factory-utilities.composition.json")
	edited := bytesReplace(t, readFile(t, path), `"version": "0.1.0"`, `"version": "0.9.0"`)
	testsupport.WriteFileIn(t, workspace.dirs.Compositions, "factory-utilities.composition.json", edited)
	after, err := Lineage(library, workspace.dirs, "factory-utilities")
	if err != nil {
		t.Fatal(err)
	}
	if composition := after.Plants[0].Composition; composition == nil || composition.Status != "out_of_sync" || after.Plants[0].Editable.Reason != "composition_out_of_sync" {
		t.Fatalf("lineage = %+v, want the edited composition out of sync", after.Plants)
	}
}

func TestLineageRecognisesAShippedExampleDomainWithoutAComposition(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	root := t.TempDir()
	copyTo(t, filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"), root, testsupport.WorkspaceDomain("chiller-pilot"))
	got, err := Lineage(library, WorkspaceDirs{Domains: testsupport.WorkspaceFolder(root, "domains")}, "")
	if err != nil || got.Plants[0].Editable.Reason != "example" || got.Plants[0].Editable.Route.Code != "create_from_example" {
		t.Fatalf("lineage = %+v, %v, want the chiller pilot recognized as the shipped example", got.Plants, err)
	}
}

func TestLineageRefusesAMissingFolderAndNeedsALibrary(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	if _, err := Lineage(library, WorkspaceDirs{Domains: filepath.Join(t.TempDir(), "none")}, ""); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Lineage of a missing domains folder = %v, want fs.ErrNotExist", err)
	}
	if _, err := Lineage(nil, WorkspaceDirs{}, ""); !errors.Is(err, ErrNoLibrary) {
		t.Fatalf("Lineage without a library = %v, want ErrNoLibrary", err)
	}
	if _, err := Lineage(&Library{}, WorkspaceDirs{}, ""); !errors.Is(err, ErrNoLibrary) {
		t.Fatalf("Lineage with an empty library = %v, want ErrNoLibrary", err)
	}
}
