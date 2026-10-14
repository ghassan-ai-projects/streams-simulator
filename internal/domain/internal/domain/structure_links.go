package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type linkIdentity struct{ linkType, from, to string }

func (checks domainChecks) checkLinks() error {
	spec := checks.compiled.Spec
	seen := map[linkIdentity]bool{}
	return checks.each("links", len(spec.Links), func(at domainChecks, i int) error {
		if err := at.linkEnds(spec.Links[i]); err != nil {
			return err
		}
		return at.linkOnce(spec.Links[i], seen)
	})
}

func (checks domainChecks) linkOnce(link model.Link, seen map[linkIdentity]bool) error {
	identity := linkIdentity{link.Type, link.From, link.To}
	if seen[identity] {
		return checks.bad("", "link %q from %q to %q appears twice", link.Type, link.From, link.To)
	}
	seen[identity] = true
	return nil
}

func (checks domainChecks) linkEnds(link model.Link) error {
	lt := checks.linkType(link.Type)
	if lt == nil {
		return checks.bad("/type", "link type %q is not declared", link.Type)
	}
	if err := checks.endpointOfKind(link.From, lt.From, "/from"); err != nil {
		return err
	}
	return checks.endpointOfKind(link.To, lt.To, "/to")
}

func (checks domainChecks) linkType(name string) *model.LinkType {
	spec := checks.compiled.Spec
	for i := range spec.LinkTypes {
		if spec.LinkTypes[i].Name == name {
			return &spec.LinkTypes[i]
		}
	}
	return nil
}

func (checks domainChecks) endpointOfKind(id, kind, member string) error {
	set := checks.compiled.InstanceSetFor(id)
	if set < 0 {
		return checks.bad(member, "instance %q is not a name or an id of any instance set", id)
	}
	if got := checks.compiled.Spec.Instances[set].Kind; got != kind {
		return checks.bad(member, "instance %q is a %s, but its link type needs a %s", id, got, kind)
	}
	return nil
}

func (checks domainChecks) checkBoards() error {
	spec := checks.compiled.Spec
	names, hosted := map[string]bool{}, map[string]string{}
	return checks.each("boards", len(spec.Boards), func(at domainChecks, i int) error {
		board := &spec.Boards[i]
		if err := registerSymbol(names, board.Name, "board", at.at+"/name", checks.source); err != nil {
			return err
		}
		return at.checkMembers(board, hosted)
	})
}

func (checks domainChecks) checkMembers(board *model.Board, hosted map[string]string) error {
	return checks.each("members", len(board.Members), func(at domainChecks, i int) error {
		member := board.Members[i]
		if at.compiled.InstanceSetFor(member) < 0 {
			return at.bad("", "board %q member %q is not a name or an id of any instance set", board.Name, member)
		}
		if owner, taken := hosted[member]; taken {
			return at.bad("", "instance %q is already on board %q, and an instance is on at most one board", member, owner)
		}
		hosted[member] = board.Name
		return nil
	})
}
