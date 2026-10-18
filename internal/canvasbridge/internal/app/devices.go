package app

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (b *Bridge) listDevices(listing *domain.Listing) error {
	kinds, err := b.filesIn("kinds")
	if err != nil {
		return err
	}
	listing.Kinds = make([]domain.KindEntry, 0, len(kinds))
	for _, rel := range kinds {
		listing.Kinds = append(listing.Kinds, domain.KindEntryOf(rel, b.listedContent(rel)))
	}
	return b.listComponents(listing)
}

func (b *Bridge) listComponents(listing *domain.Listing) error {
	listing.Components = []domain.ComponentEntry{}
	for _, folder := range domain.ComponentFolders() {
		class, _, _ := domain.ClassOfFolder(folder)
		rels, err := b.filesIn(folder)
		if err != nil {
			return err
		}
		for _, rel := range rels {
			listing.Components = append(listing.Components, domain.ComponentEntryOf(rel, class, b.listedContent(rel)))
		}
	}
	return nil
}

func (b *Bridge) filesIn(folder string) ([]string, error) {
	_, suffix, _ := domain.ClassOfFolder(folder)
	names, err := b.edges.Workspace.Names(folder)
	if err != nil {
		return nil, err
	}
	rels := []string{}
	for _, name := range names {
		if strings.HasSuffix(name, suffix) && !strings.HasPrefix(name, ".") {
			rels = append(rels, folder+"/"+name)
		}
	}
	return rels, nil
}

func (b *Bridge) content(rel string) []byte {
	return b.readUpTo(rel, domain.MaxFileBytes)
}

func (b *Bridge) listedContent(rel string) []byte {
	return b.readUpTo(rel, domain.MaxListedFileBytes)
}

func (b *Bridge) readUpTo(rel string, limit int64) []byte {
	content, err := b.edges.Workspace.Read(rel, limit)
	if err != nil {
		return nil
	}
	return content
}
