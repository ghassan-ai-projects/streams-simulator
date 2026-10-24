package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const (
	maxRunFile   = 32 << 20
	observerHead = 64 << 10
)

type WorkspaceDirs struct {
	Domains, Kinds, Twins, Compositions, Runs string
}

type folder struct {
	dir, label, suffix string
	into               *[]domain.NamedFile
}

func foldersOf(dirs WorkspaceDirs, workspace *domain.Workspace) []folder {
	return []folder{
		{dirs.Domains, "domains", ".domain.json", &workspace.Domains}, {dirs.Kinds, "kinds", ".kind.json", &workspace.Kinds},
		{dirs.Twins, "twins", ".twin.json", &workspace.Twins}, {dirs.Compositions, "compositions", ".composition.json", &workspace.Compositions},
	}
}

func Lineage(lib *domain.Library, dirs WorkspaceDirs, only string) (*domain.Lineage, error) {
	workspace, err := ReadWorkspace(dirs)
	if err != nil {
		return nil, err
	}
	return domain.DescribeLineage(lib, workspace, only)
}

func ReadWorkspace(dirs WorkspaceDirs) (domain.Workspace, error) {
	var workspace domain.Workspace
	for _, folder := range foldersOf(dirs, &workspace) {
		files, err := readNamed(folder.dir, folder.label, folder.suffix)
		if err != nil {
			return domain.Workspace{}, err
		}
		*folder.into = files
	}
	return withExamplesAndRuns(workspace, dirs.Runs)
}

func withExamplesAndRuns(workspace domain.Workspace, runsDir string) (domain.Workspace, error) {
	examples, examplesErr := readGroup(builtInExampleTree, ".", "examples", ".composition.json")
	runs, runsErr := readRuns(runsDir)
	for _, example := range examples {
		workspace.Examples = append(workspace.Examples, domain.NamedFile{Name: path.Base(example.Path), Raw: example.Raw})
	}
	workspace.Runs = runs
	return workspace, errors.Join(examplesErr, runsErr)
}

func readNamed(dir, label, suffix string) ([]domain.NamedFile, error) {
	if dir == "" {
		return nil, nil
	}
	names, err := kernel.Names(dir)
	if err != nil {
		return nil, fmt.Errorf("component: list %s: %w", dir, err)
	}
	var files []domain.NamedFile
	for _, name := range sortedWithSuffix(names, suffix) {
		raw, err := kernel.Read(filepath.Join(dir, name))
		files = append(files, domain.NamedFile{Name: name, Raw: raw, Err: labeled(err, label+"/"+name)})
	}
	return files, nil
}

func labeled(err error, label string) error {
	if err == nil {
		return nil
	}
	return &fs.PathError{Op: "read", Path: label, Err: causeOf(err)}
}

func causeOf(err error) error {
	for _, sentinel := range []error{kernel.ErrNotRegular, kernel.ErrTooLarge} {
		if errors.Is(err, sentinel) {
			return sentinel
		}
	}
	cause := err
	var pathError *fs.PathError
	if errors.As(err, &pathError) {
		cause = pathError.Err
	}
	return cause
}

func readRuns(dir string) ([]domain.RunFolder, error) {
	if dir == "" {
		return nil, nil
	}
	ids, err := kernel.DirectoryNames(dir)
	if err != nil {
		return nil, fmt.Errorf("component: list %s: %w", dir, err)
	}
	var runs []domain.RunFolder
	for _, id := range ids {
		runs = append(runs, readRun(dir, id))
	}
	return runs, nil
}

func readRun(dir, id string) domain.RunFolder {
	folder := domain.RunFolder{ID: id}
	raw, err := kernel.ReadRegular(filepath.Join(dir, id, "run.json"), maxRunFile)
	switch {
	case err == nil:
		folder.RunJSON = raw
	case errors.Is(err, fs.ErrNotExist):
		folder.Observer, err = readObserverHead(filepath.Join(dir, id, "journal.jsonl"))
		folder.Err = labeled(err, "runs/"+id+"/journal.jsonl")
	default:
		folder.Err = labeled(err, "runs/"+id+"/run.json")
	}
	return folder
}

func readObserverHead(file string) ([]byte, error) {
	journal, present, err := kernel.OpenIfPresent(file)
	if err != nil || !present {
		return nil, err
	}
	head, err := io.ReadAll(io.LimitReader(journal, observerHead))
	return head, errors.Join(err, journal.Close())
}
