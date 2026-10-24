package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Relation struct {
	Type     string   `json:"type"`
	FromKind string   `json:"from_kind"`
	ToKind   string   `json:"to_kind"`
	Links    []string `json:"links"`
}

func relationsOf(spec *model.DomainSpec, kind string) []Relation {
	relations := []Relation{}
	for _, linkType := range spec.LinkTypes {
		if linkType.From == kind || linkType.To == kind {
			relations = append(relations, Relation{Type: linkType.Name, FromKind: linkType.From, ToKind: linkType.To, Links: linksOfType(spec, linkType.Name)})
		}
	}
	return relations
}

func linksOfType(spec *model.DomainSpec, linkType string) []string {
	links := []string{}
	for _, link := range spec.Links {
		if link.Type == linkType {
			links = append(links, fmt.Sprintf("%s>%s", link.From, link.To))
		}
	}
	return links
}
