package app

import (
	"bytes"
	"strings"
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
	_, suffix, _ := domain.ClassOfFolder(folder)
	names, err := b.edges.Workspace.Names(folder)
	if err != nil {
		return nil, err
	}
	out := []domain.Definition{}
	for _, name := range names {
		if strings.HasSuffix(name, suffix) && !strings.HasPrefix(name, ".") {
			out = append(out, b.definition(folder+"/"+name))
		}
	}
	return out, nil
}

func (b *Bridge) definition(rel string) domain.Definition {
	content, err := b.edges.Workspace.Read(rel, domain.MaxFileBytes)
	if err != nil {
		return domain.Definition{Path: rel}
	}
	return domain.DefinitionOf(rel, content)
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
	summary := b.summary(name)
	return domain.RunEntry{
		WorldID:    name,
		Ended:      summary.State(b.runIsRunning(name)).Ended,
		Explore:    summary.Explore,
		Plant:      summary.Plant,
		DurationNs: summary.DurationNs,
		Faults:     b.faults(name),
		RecordedAt: b.recordedAt("runs/" + name + "/journal.jsonl"),
	}
}

func (b *Bridge) faults(runID string) int {
	content, err := b.edges.Workspace.Read("runs/"+runID+"/run.json", domain.MaxFileBytes)
	if err != nil {
		return 0
	}
	return domain.FaultsInjected(content)
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
