package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func fileNames(files []domain.NamedFile) []string {
	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name
	}
	return names
}

func TestReadWorkspaceReadsEachFolderByItsSuffixAndTheEmbeddedExamples(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for rel, content := range map[string]string{
		testsupport.WorkspaceDomain("a"): "A", testsupport.WorkspaceFile("domains", "note.txt"): "x", "kinds/fan.kind.json": "F", "twins/t.twin.json": "T",
		"compositions/c.composition.json": "C", "runs/r-1/run.json": `{"domain":{}}`, "runs/r-2/journal.jsonl": "first\nsecond", "runs/r-3/ledger.jsonl": "x",
	} {
		write(t, filepath.Join(root, rel), content)
	}
	dirs := WorkspaceDirs{Domains: testsupport.WorkspaceFolder(root, "domains"), Kinds: filepath.Join(root, "kinds"), Twins: filepath.Join(root, "twins"),
		Compositions: filepath.Join(root, "compositions"), Runs: filepath.Join(root, "runs")}
	workspace, err := ReadWorkspace(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileNames(workspace.Domains); len(got) != 1 || got[0] != "a.domain.json" || string(workspace.Domains[0].Raw) != "A" {
		t.Fatalf("domains = %v, want a.domain.json only", got)
	}
	if len(workspace.Kinds) != 1 || len(workspace.Twins) != 1 || len(workspace.Compositions) != 1 {
		t.Fatalf("workspace = %+v, want one kind, twin and composition", workspace)
	}
	if got := fileNames(workspace.Examples); len(got) != 2 || got[0] != "chiller-pilot.composition.json" {
		t.Fatalf("examples = %v, want the shipped example compositions", got)
	}
	runs := workspace.Runs
	if len(runs) != 3 || string(runs[0].RunJSON) != `{"domain":{}}` || runs[0].Observer != nil || runs[1].RunJSON != nil || string(runs[1].Observer) != "first\nsecond" || runs[2].RunJSON != nil || runs[2].Observer != nil || runs[2].Err != nil {
		t.Fatalf("runs = %+v, want run.json, the observer journal head, and a folder with neither", runs)
	}
}

func TestReadWorkspaceSkipsAFolderThatWasNotGiven(t *testing.T) {
	t.Parallel()
	workspace, err := ReadWorkspace(WorkspaceDirs{})
	if err != nil || workspace.Domains != nil || workspace.Kinds != nil || workspace.Runs != nil || len(workspace.Examples) == 0 {
		t.Fatalf("ReadWorkspace = %+v, %v, want nothing but the examples", workspace, err)
	}
}

func TestReadWorkspaceRefusesAFolderThatIsNotThere(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, dirs := range map[string]WorkspaceDirs{
		"domains": {Domains: filepath.Join(root, "none")}, "runs": {Runs: filepath.Join(root, "none")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := ReadWorkspace(dirs); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("ReadWorkspace = %v, want fs.ErrNotExist", err)
			}
		})
	}
}

func TestReadWorkspaceCarriesAFileItCannotReadWithoutThePathOfTheMachine(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "kinds")
	write(t, filepath.Join(dir, "ok.kind.json"), "K")
	if err := os.Symlink(filepath.Join(root, "nowhere"), filepath.Join(dir, "gone.kind.json")); err != nil {
		t.Fatal(err)
	}
	workspace, err := ReadWorkspace(WorkspaceDirs{Kinds: dir})
	if err != nil || len(workspace.Kinds) != 2 {
		t.Fatalf("ReadWorkspace = %+v, %v, want both kind files, one with its error", workspace.Kinds, err)
	}
	gone := workspace.Kinds[0]
	if gone.Name != "gone.kind.json" || !errors.Is(gone.Err, fs.ErrNotExist) || strings.Contains(gone.Err.Error(), root) || !strings.Contains(gone.Err.Error(), "kinds/gone.kind.json") {
		t.Fatalf("gone = %+v, want a not-exist error labeled kinds/gone.kind.json without %s", gone, root)
	}
}

func TestReadWorkspaceCarriesARunFolderItCannotReadAsAnError(t *testing.T) {
	t.Parallel()
	runs := filepath.Join(t.TempDir(), "runs")
	write(t, filepath.Join(runs, "r-dir", "run.json", "x"), "x")
	write(t, filepath.Join(runs, "r-journal", "journal.jsonl", "x"), "x")
	big := filepath.Join(runs, "r-big", "run.json")
	write(t, big, "")
	if err := os.Truncate(big, maxRunFile+1); err != nil {
		t.Fatal(err)
	}
	workspace, err := ReadWorkspace(WorkspaceDirs{Runs: runs})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]error{}
	for _, run := range workspace.Runs {
		byID[run.ID] = run.Err
	}
	for id, want := range map[string]error{"r-dir": kernel.ErrNotRegular, "r-big": kernel.ErrTooLarge} {
		if !errors.Is(byID[id], want) || strings.Contains(byID[id].Error(), runs) {
			t.Errorf("%s error = %v, want %v without the machine path", id, byID[id], want)
		}
	}
	var pathError *fs.PathError
	if !errors.As(byID["r-journal"], &pathError) || pathError.Path != "runs/r-journal/journal.jsonl" {
		t.Errorf("r-journal error = %v, want a path error labeled runs/r-journal/journal.jsonl", byID["r-journal"])
	}
}

func TestLineageReadsTheWorkspaceAndRefusesAFolderThatIsNotThere(t *testing.T) {
	t.Parallel()
	library, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "kinds", "fan.kind.json"), "{")
	lineage, err := Lineage(library, WorkspaceDirs{Kinds: filepath.Join(root, "kinds")}, "")
	if err != nil || len(lineage.Plants) != 0 || len(lineage.Unreadable) != 1 || lineage.Unreadable[0].Code != domain.UnreadableKind {
		t.Fatalf("Lineage = %+v, %v, want no plant and the broken kind file named", lineage, err)
	}
	if _, err := Lineage(library, WorkspaceDirs{Kinds: filepath.Join(root, "none")}, ""); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Lineage of a missing folder = %v, want fs.ErrNotExist", err)
	}
}
