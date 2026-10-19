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
	ClassKind        FileClass = "kind"
	ClassTwin        FileClass = "twin"
	ClassEquipment   FileClass = "equipment"
	ClassInstrument  FileClass = "instrument"
	ClassRun         FileClass = "run"
	ClassPolicies    FileClass = "policies"
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
	"domains":                {ClassDomain, ".domain.json"},
	"compositions":           {ClassComposition, ".composition.json"},
	"scenarios":              {ClassScenario, ".scenario.json"},
	"layouts":                {ClassLayout, ".layout.json"},
	"kinds":                  {ClassKind, ".kind.json"},
	"twins":                  {ClassTwin, ".twin.json"},
	"components/equipment":   {ClassEquipment, ".component.json"},
	"components/instruments": {ClassInstrument, ".instrument.json"},
}

const policyDefaultRel = "settings/default.policies.json"

var stemFolders = map[FileClass]string{ClassKind: "kinds"}

var runFiles = []string{"run.json", "journal.jsonl", "trace.jsonl", "ledger.jsonl", "world_state_history.jsonl",
	"field-trace.jsonl", "interrupts.jsonl", "verdict.json", "label.json", "sealed/journal.jsonl"}

func ParsePath(rel string) (WorkspacePath, error) {
	parts := strings.Split(rel, "/")
	if !slices.ContainsFunc(parts, badSegment) {
		if path, ok := knownPath(rel, parts); ok {
			return path, nil
		}
	}
	return WorkspacePath{}, fmt.Errorf("%w: %q", ErrBadPath, rel)
}

func knownPath(rel string, parts []string) (WorkspacePath, bool) {
	if path, ok := definitionPath(rel, parts); ok {
		return path, true
	}
	if path, ok := runPath(rel, parts); ok {
		return path, true
	}
	return policyDefaultPath(rel)
}

func policyDefaultPath(rel string) (WorkspacePath, bool) {
	return WorkspacePath{Rel: rel, Class: ClassPolicies}, rel == policyDefaultRel
}

func StemPath(class FileClass, stem string) (WorkspacePath, error) {
	folder, named := stemFolders[class]
	if !named {
		return WorkspacePath{}, fmt.Errorf("%w: no folder holds %s files by name", ErrBadPath, class)
	}
	path, err := ParsePath(folder + "/" + stem + definitionFolders[folder].suffix)
	if err != nil || path.Class != class || strings.Contains(stem, "/") {
		return WorkspacePath{}, fmt.Errorf("want the name of a %s file without folder or suffix, got %q", class, stem)
	}
	return path, nil
}

func RunJournal(runID string) (WorkspacePath, error) {
	if badSegment(runID) {
		return WorkspacePath{}, fmt.Errorf("%w: run id %q", ErrBadPath, runID)
	}
	return WorkspacePath{Rel: "runs/" + runID + "/journal.jsonl", Class: ClassRun, RunID: runID, RunFile: "journal.jsonl"}, nil
}

func badSegment(part string) bool {
	return !segment.MatchString(part) || part == "." || part == ".."
}

func definitionPath(rel string, parts []string) (WorkspacePath, bool) {
	name := parts[len(parts)-1]
	folder, known := definitionFolders[strings.Join(parts[:len(parts)-1], "/")]
	if !known || !strings.HasSuffix(name, folder.suffix) || name == folder.suffix {
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

func ComponentFolders() []string {
	return []string{"components/equipment", "components/instruments"}
}

func ClassOfFolder(folder string) (FileClass, string, bool) {
	entry, known := definitionFolders[folder]
	return entry.class, entry.suffix, known
}
