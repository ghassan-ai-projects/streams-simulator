package app

import (
	"encoding/json"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const firstLineLimit = 1 << 20

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
	entry := domain.Definition{Path: rel}
	content, err := b.edges.Workspace.Read(rel, domain.MaxFileBytes)
	if err != nil {
		return entry
	}
	var head struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if json.Unmarshal(content, &head) == nil {
		entry.ID, entry.Title = head.ID, head.Title
	}
	return entry
}

func (b *Bridge) runs() ([]domain.RunEntry, error) {
	names, err := b.edges.Workspace.Names("runs")
	if err != nil {
		return nil, err
	}
	runs := []domain.RunEntry{}
	for _, name := range names {
		if b.edges.Workspace.Exists("runs/" + name + "/journal.jsonl") {
			state := b.runState(name)
			runs = append(runs, domain.RunEntry{WorldID: name, Ended: state.Ended, Explore: state.Explore})
		}
	}
	return runs, nil
}
