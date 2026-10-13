package domain

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type FileClass string

const (
	ClassDomain      FileClass = "domain"
	ClassComposition FileClass = "composition"
	ClassScenario    FileClass = "scenario"
	ClassLayout      FileClass = "layout"
	ClassRun         FileClass = "run"
)

const MaxFileBytes = 8 << 20

type WorkspacePath struct {
	Rel     string
	Class   FileClass
	RunID   string
	RunFile string
}

var segment = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

var definitionFolders = map[string]struct {
	class  FileClass
	suffix string
}{
	"domains":      {ClassDomain, ".domain.json"},
	"compositions": {ClassComposition, ".composition.json"},
	"scenarios":    {ClassScenario, ".scenario.json"},
	"layouts":      {ClassLayout, ".layout.json"},
}

var runFiles = []string{"run.json", "journal.jsonl", "trace.jsonl", "ledger.jsonl", "world_state_history.jsonl",
	"verdict.json", "label.json", "sealed/journal.jsonl"}

func ParsePath(rel string) (WorkspacePath, error) {
	parts := strings.Split(rel, "/")
	if !slices.ContainsFunc(parts, badSegment) {
		if path, ok := definitionPath(rel, parts); ok {
			return path, nil
		}
		if path, ok := runPath(rel, parts); ok {
			return path, nil
		}
	}
	return WorkspacePath{}, fmt.Errorf("%w: %q", ErrBadPath, rel)
}

func badSegment(part string) bool {
	return !segment.MatchString(part) || part == "." || part == ".."
}

func definitionPath(rel string, parts []string) (WorkspacePath, bool) {
	folder, known := definitionFolders[parts[0]]
	if !known || len(parts) != 2 || !strings.HasSuffix(parts[1], folder.suffix) || parts[1] == folder.suffix {
		return WorkspacePath{}, false
	}
	return WorkspacePath{Rel: rel, Class: folder.class}, true
}

func runPath(rel string, parts []string) (WorkspacePath, bool) {
	if len(parts) < 3 || parts[0] != "runs" {
		return WorkspacePath{}, false
	}
	file := strings.Join(parts[2:], "/")
	if !slices.Contains(runFiles, file) {
		return WorkspacePath{}, false
	}
	return WorkspacePath{Rel: rel, Class: ClassRun, RunID: parts[1], RunFile: file}, true
}

func (p WorkspacePath) Writable() error {
	if p.Class == ClassRun {
		return fmt.Errorf("%w: %s", ErrReadOnly, p.Rel)
	}
	return nil
}

func DefinitionFolders() []string {
	return []string{"domains", "compositions", "scenarios", "layouts"}
}

func ClassOfFolder(folder string) (FileClass, string, bool) {
	entry, known := definitionFolders[folder]
	return entry.class, entry.suffix, known
}
