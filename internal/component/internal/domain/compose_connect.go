package domain

import (
	"strings"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

type linkSet struct {
	types   []any
	links   []any
	seen    map[string]bool
	follows []follow
}

type follow struct {
	kind, state, link, source string
	sampleS                   float64
}

type endpoint struct {
	owner    *built
	instance string
	port     Port
}

func (c *composer) connect(entries []*built) (*linkSet, error) {
	set := &linkSet{seen: map[string]bool{}}
	for index, connection := range c.composition.Connections {
		if err := c.connectOne(entries, set, index, connection); err != nil {
			return nil, err
		}
	}
	return set, c.applyFollows(entries, set)
}

func (c *composer) connectOne(entries []*built, set *linkSet, index int, connection Connection) error {
	at := pointer("connections", indexPart(index))
	from, err := c.endpoint(entries, connection.From, "provides", at+"/from")
	if err != nil {
		return err
	}
	to, err := c.endpoint(entries, connection.To, "consumes", at+"/to")
	if err != nil {
		return err
	}
	if from.port.Type != to.port.Type {
		return c.refuse(at, "port %s.%s is of type %q, port %s.%s of type %q", from.instance, from.port.Name, from.port.Type, to.instance, to.port.Name, to.port.Type)
	}
	return set.add(c, from, to, at)
}

func (c *composer) endpoint(entries []*built, text, direction, at string) (endpoint, error) {
	cut := strings.LastIndex(text, ".")
	instance, portName := text[:cut], text[cut+1:]
	owner := findInstance(entries, instance)
	if owner == nil {
		return endpoint{}, c.refuse(at, "%q names no instance of the composition", instance)
	}
	port, err := c.portOf(owner, instance, portName, direction, at)
	return endpoint{owner: owner, instance: instance, port: port}, err
}

func (c *composer) portOf(owner *built, instance, name, direction, at string) (Port, error) {
	port, found := owner.ports[name]
	switch {
	case !found:
		return Port{}, c.refuse(at, "instance %q (kind %s) has no port %q", instance, owner.kindName, name)
	case port.Direction != direction:
		return Port{}, c.refuse(at, "port %q of %q %s; this end of a connection must be a port that %s", name, instance, port.Direction, direction)
	}
	return port, nil
}

func findInstance(entries []*built, instance string) *built {
	for _, entry := range entries {
		if specs.InstanceRenders(entry.entry.instances(), instance) {
			return entry
		}
	}
	return nil
}

func (s *linkSet) add(c *composer, from, to endpoint, at string) error {
	name := to.owner.kindName + "__" + to.port.Name + "__" + from.owner.kindName
	if !s.seen[name] {
		s.seen[name] = true
		s.types = append(s.types, map[string]any{"name": name, "from": from.owner.kindName, "to": to.owner.kindName})
	}
	s.links = append(s.links, map[string]any{"type": name, "from": from.instance, "to": to.instance})
	return s.record(c, follow{kind: to.owner.kindName, state: to.port.State, link: name, source: from.port.State, sampleS: to.port.SampleS}, at)
}

func (s *linkSet) record(c *composer, wanted follow, at string) error {
	for _, existing := range s.follows {
		if existing.kind != wanted.kind || existing.state != wanted.state {
			continue
		}
		if existing != wanted {
			return c.refuse(at, "state %q of kind %s is fed by %s reading %q and by %s reading %q: one link type and one provided state per consuming state",
				wanted.state, wanted.kind, existing.link, existing.source, wanted.link, wanted.source)
		}
		return nil
	}
	s.follows = append(s.follows, wanted)
	return nil
}
