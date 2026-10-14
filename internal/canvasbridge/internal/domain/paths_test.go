package domain

import (
	"errors"
	"testing"
)

func inDomains(name string) string {
	return DefinitionFolders()[0] + "/" + name
}

func TestOnlyTheWorkspaceFilesOfTheCanvasAreAccepted(t *testing.T) {
	t.Parallel()
	good := map[string]WorkspacePath{
		inDomains("plant.domain.json"):    {Rel: inDomains("plant.domain.json"), Class: ClassDomain},
		"compositions/f.composition.json": {Rel: "compositions/f.composition.json", Class: ClassComposition},
		"scenarios/s.scenario.json":       {Rel: "scenarios/s.scenario.json", Class: ClassScenario},
		"layouts/plant.layout.json":       {Rel: "layouts/plant.layout.json", Class: ClassLayout},
		"runs/w-1/run.json":               {Rel: "runs/w-1/run.json", Class: ClassRun, RunID: "w-1", RunFile: "run.json"},
		"runs/w-1/sealed/journal.jsonl":   {Rel: "runs/w-1/sealed/journal.jsonl", Class: ClassRun, RunID: "w-1", RunFile: "sealed/journal.jsonl"},
	}
	for rel, want := range good {
		if got, err := ParsePath(rel); err != nil || got != want {
			t.Errorf("ParsePath(%q) = %+v, %v; want %+v", rel, got, err, want)
		}
	}
	for _, rel := range []string{"", "/etc/passwd", "../" + inDomains("x.domain.json"), inDomains("../runs/w/run.json"), inDomains("x.json"),
		inDomains(".domain.json"), inDomains("a/b.domain.json"), inDomains(".hidden.domain.json"), "runs/w-1/secret.txt", "runs/../run.json",
		"runs/w-1/sealed/other.jsonl", `domains\x.domain.json`, inDomains("/x.domain.json"), "Domains/x.domain.json"} {
		if _, err := ParsePath(rel); !errors.Is(err, ErrBadPath) {
			t.Errorf("ParsePath(%q) = %v, want ErrBadPath", rel, err)
		}
	}
}

func TestRunFilesAreReadOnlyAndDefinitionsWritable(t *testing.T) {
	t.Parallel()
	run, _ := ParsePath("runs/w-1/run.json")
	domain, _ := ParsePath(inDomains("p.domain.json"))
	if !errors.Is(run.Writable(), ErrReadOnly) || domain.Writable() != nil {
		t.Fatal("runs are evidence: read only; definitions are editable")
	}
	if class, suffix, ok := ClassOfFolder("layouts"); !ok || class != ClassLayout || suffix != ".layout.json" || len(DefinitionFolders()) != 4 {
		t.Fatal("definition folders")
	}
}

func TestSealedRunFilesWaitForTheEndUnlessTheRunIsExplore(t *testing.T) {
	t.Parallel()
	open, ended, explore := RunState{}, RunState{Ended: true}, RunState{Explore: true}
	for file, sealedWhileOpen := range map[string]bool{"journal.jsonl": false, "trace.jsonl": false, "run.json": true,
		"sealed/journal.jsonl": true, "ledger.jsonl": true, "world_state_history.jsonl": true} {
		path, err := ParsePath("runs/w-1/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if got := errors.Is(ReadableRunFile(path, open), ErrSealed); got != sealedWhileOpen {
			t.Errorf("%s while the Evaluate run is open: sealed %v, want %v", file, got, sealedWhileOpen)
		}
		if ReadableRunFile(path, ended) != nil || ReadableRunFile(path, explore) != nil {
			t.Errorf("%s is readable once the run ended or in Explore", file)
		}
	}
}
