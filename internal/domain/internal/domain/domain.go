// Package domain holds the domain-spec rules: parsing, schema validation,
// cross-checking, compilation and digesting of the data files that define
// whole simulated worlds, and the catalog over them. The binary contains no
// domain behavior: a domain that needs a code branch in the binary is a bug
// in the simulator's design, not a missing feature. It performs no I/O; the
// files edge reads the bytes.
package domain

import (
	"fmt"
	"regexp"
	"slices"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Parse reads, schema-validates and structurally validates a domain spec from
// an in-memory document. The returned Compiled carries the spec, its canonical
// digest, and resolved defaults. src names the source for error messages.
func Parse(raw []byte, src string) (*Compiled, error) {
	doc, err := validateDocument(raw, src)
	if err != nil {
		return nil, err
	}
	return compileDocument(raw, doc, src)
}

// Compiled is the typed, validated, digest-carrying form of a domain spec.
type Compiled struct {
	Spec   *model.DomainSpec
	Digest string
	// Raw is the validated source document embedded in run artifacts.
	Raw []byte

	channels  map[string]bool
	faults    map[string]bool
	effectors map[string]bool
	profiles  map[string]bool

	idPatterns []*regexp.Regexp
	names      []map[string]bool
	effective  []*model.Kind
	lookup     []*model.Kind

	// Resolved defaults (schema defaults applied).
	channelGain map[string]float64 // channel -> observation_gain
}

func (c *Compiled) HasChannel(name string) bool { return c.channels[name] }

func (c *Compiled) HasFault(id string) bool { return c.faults[id] }

func (c *Compiled) HasEffector(name string) bool { return c.effectors[name] }

func (c *Compiled) HasProfile(name string) bool { return c.profiles[name] }

func (c *Compiled) ChannelGain(ch string) float64 {
	if g, ok := c.channelGain[ch]; ok {
		return g
	}
	return 1
}

func (c *Compiled) Kind(name string) *model.Kind {
	for i := range c.Spec.Kinds {
		if c.Spec.Kinds[i].Name == name {
			return &c.Spec.Kinds[i]
		}
	}
	return nil
}

func (c *Compiled) StateNames(kind string) []string {
	var out []string
	if k := c.Kind(kind); k != nil {
		for i := range k.State {
			out = append(out, k.State[i].Name)
		}
	}
	sort.Strings(out)
	return out
}

func (c *Compiled) ChannelNames() []string { return sortedKeys(c.channels) }

func (c *Compiled) KindChannelNames(kind string) []string {
	var out []string
	for _, k := range c.lookup {
		if k.Name == kind {
			out = append(out, channelNames(k)...)
		}
	}
	return slices.Compact(slices.Sorted(slices.Values(out)))
}

func channelNames(kind *model.Kind) []string {
	out := make([]string, len(kind.Channels))
	for i := range kind.Channels {
		out[i] = kind.Channels[i].Name
	}
	return out
}

func (c *Compiled) FaultCount() int { return len(c.faults) }

func (c *Compiled) EffectorCount() int { return len(c.effectors) }

func (c *Compiled) Fault(id string) *model.Fault {
	for _, kind := range c.lookup {
		if f := kind.FaultDecl(id); f != nil {
			return f
		}
	}
	return nil
}

func (c *Compiled) FaultKind(id string) *model.Kind {
	for _, kind := range c.lookup {
		if kind.FaultDecl(id) != nil {
			return kind
		}
	}
	return nil
}

func (c *Compiled) Effector(name string) *model.Effector {
	for _, kind := range c.lookup {
		if e := kind.EffectorDecl(name); e != nil {
			return e
		}
	}
	return nil
}

func (c *Compiled) Channel(name string) *model.Channel {
	for _, kind := range c.lookup {
		if ch := kind.ChannelDecl(name); ch != nil {
			return ch
		}
	}
	return nil
}

func (c *Compiled) Profile(name string) *model.Profile {
	for i := range c.Spec.Profiles {
		if c.Spec.Profiles[i].Name == name {
			return &c.Spec.Profiles[i]
		}
	}
	return nil
}

func compile(spec *model.DomainSpec, digest, src string) (*Compiled, error) {
	c := newCompiled(spec, digest)
	if err := c.checkInstanceKinds(src); err != nil {
		return nil, err
	}
	c.compileInstanceIDs()
	if err := c.registerSymbols(src); err != nil {
		return nil, err
	}
	if err := crossCheck(c, src); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	c.indexLookup()
	return c, nil
}
