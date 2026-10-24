package domain

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var coveredDomainMembers = []string{"id", "version", "kinds", "instances", "links", "link_types"}

func diffDomain(a, b *diffSide, tally *UnchangedSummary) []Change {
	links, linkUnchanged := diffGroup("", linkGroup(a.spec.Links), linkGroup(b.spec.Links))
	types, typeUnchanged := diffGroup("", linkTypeGroup(a.spec.LinkTypes), linkTypeGroup(b.spec.LinkTypes))
	tally.Links, tally.LinkTypes = linkUnchanged, typeUnchanged
	return slices.Concat(links, types, diffTopLevel(a.spec, b.spec))
}

func linkGroup(links []model.Link) group {
	shape := group{noun: "link", segment: SectionLinks, section: SectionLinks}
	for _, link := range links {
		shape.entries = append(shape.entries, newEntry(linkKey(link), link, link, nil))
	}
	return shape
}

func linkKey(link model.Link) string {
	from, to := link.From, link.To
	if link.FromPort != "" {
		from += "." + link.FromPort
	}
	if link.ToPort != "" {
		to += "." + link.ToPort
	}
	return fmt.Sprintf("%s:%s>%s", link.Type, from, to)
}

func linkTypeGroup(types []model.LinkType) group {
	shape := group{noun: "link_type", segment: SectionLinkTypes, section: SectionLinkTypes}
	for _, one := range types {
		shape.entries = append(shape.entries, newEntry(one.Name, one, one, nil))
	}
	return shape
}

func diffTopLevel(a, b *model.DomainSpec) []Change {
	return memberChanges(genericMap(a), genericMap(b), coveredDomainMembers, CodeSectionChanged, SectionDomain, SectionDomain+"/")
}

func memberChanges(left, right map[string]any, covered []string, code, section, subject string) []Change {
	for _, member := range covered {
		delete(left, member)
		delete(right, member)
	}
	var changes []Change
	for _, member := range unionKeys(left, right) {
		if !reflect.DeepEqual(left[member], right[member]) {
			changes = append(changes, newChange(code, subject+member, section, member, left[member], right[member]))
		}
	}
	return changes
}
