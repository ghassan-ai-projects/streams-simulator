package app

import (
	"bytes"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	firstLineLimit = 1 << 20
	closingLimit   = 64 << 10
)

func (b *Bridge) Workspace() (domain.Listing, error) {
	listing := domain.Listing{Name: b.edges.Workspace.Name()}
	if err := b.listDefinitions(&listing); err != nil {
		return domain.Listing{}, err
	}
	if err := b.listDevices(&listing); err != nil {
		return domain.Listing{}, err
	}
	var err error
	listing.Runs, err = b.runs()
	return listing, err
}

func (b *Bridge) listDefinitions(listing *domain.Listing) error {
	for folder, target := range map[string]*[]domain.Definition{"domains": &listing.Domains, "compositions": &listing.Compositions,
		"scenarios": &listing.Scenarios, "layouts": &listing.Layouts} { // determinism-safe: each folder fills its own field
		definitions, err := b.definitions(folder)
		if err != nil {
			return err
		}
		*target = definitions
	}
	return nil
}

func (b *Bridge) definitions(folder string) ([]domain.Definition, error) {
	rels, err := b.filesIn(folder)
	if err != nil {
		return nil, err
	}
	out := []domain.Definition{}
	for _, rel := range rels {
		out = append(out, domain.DefinitionOf(rel, b.content(rel)))
	}
	return out, nil
}

func (b *Bridge) runs() ([]domain.RunEntry, error) {
	names, err := b.edges.Workspace.Names("runs")
	if err != nil {
		return nil, err
	}
	runs := []domain.RunEntry{}
	for _, name := range names {
		if b.edges.Workspace.Exists("runs/" + name + "/journal.jsonl") {
			runs = append(runs, b.runEntry(name))
		}
	}
	return runs, nil
}

func (b *Bridge) runEntry(name string) domain.RunEntry {
	running := b.runIsRunning(name)
	entry := domain.NewRunEntry(name, b.summary(name), running, b.artifact(name))
	entry.RecordedAt = b.recordedAt("runs/" + name + "/journal.jsonl")
	if plan, started := b.startedSession(name); started {
		return entry.AsSession(running, plan)
	}
	return b.recordedSession(name, entry, running)
}

func (b *Bridge) recordedSession(name string, entry domain.RunEntry, running bool) domain.RunEntry {
	head, err := b.edges.Workspace.ReadFrom("runs/"+name+"/journal.jsonl", 0, firstLineLimit)
	if err == nil {
		if plan, connected := domain.RecordedSession(head); connected {
			return entry.AsSession(running, plan)
		}
	}
	return entry
}

func (b *Bridge) artifact(runID string) domain.RunArtifact {
	content, err := b.edges.Workspace.Read("runs/"+runID+"/run.json", domain.MaxFileBytes)
	if err != nil {
		return domain.RunArtifact{}
	}
	return domain.ParseRunArtifact(content)
}

func (b *Bridge) summary(runID string) domain.RunSummary {
	journal := "runs/" + runID + "/journal.jsonl"
	head, headErr := b.edges.Workspace.ReadFrom(journal, 0, firstLineLimit)
	tail, tailErr := b.edges.Workspace.Tail(journal, closingLimit)
	if headErr != nil || tailErr != nil {
		return domain.RunSummary{}
	}
	first, _, _ := bytes.Cut(head, []byte("\n"))
	return domain.SummarizeRun(first, tail)
}

func (b *Bridge) recordedAt(rel string) string {
	modified, err := b.edges.Workspace.Modified(rel)
	if err != nil {
		return ""
	}
	return modified.UTC().Format(time.RFC3339)
}
