// Package domain loads, validates and compiles domain specs — the data
// files that define whole simulated worlds. The binary contains no domain
// behavior: every domain in the catalog loads through this one path, and a
// domain that needs a code branch in the binary is a bug in the simulator's
// design, not a missing feature.
package domain

import (
	"fmt"

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

	states    map[string]bool
	channels  map[string]bool
	faults    map[string]bool
	effectors map[string]bool
	profiles  map[string]bool

	// Resolved defaults (schema defaults applied).
	channelGain map[string]float64 // channel -> observation_gain
}

// HasState reports whether the state name is declared.
func (c *Compiled) HasState(name string) bool { return c.states[name] }

// HasChannel reports whether the channel name is declared.
func (c *Compiled) HasChannel(name string) bool { return c.channels[name] }

// HasFault reports whether the fault id is declared.
func (c *Compiled) HasFault(id string) bool { return c.faults[id] }

// HasEffector reports whether the effector name is declared.
func (c *Compiled) HasEffector(name string) bool { return c.effectors[name] }

// HasProfile reports whether the profile name is declared.
func (c *Compiled) HasProfile(name string) bool { return c.profiles[name] }

// ChannelGain returns the resolved observation gain for a channel.
func (c *Compiled) ChannelGain(ch string) float64 {
	if g, ok := c.channelGain[ch]; ok {
		return g
	}
	return 1
}

// StateNames returns the declared state names, sorted.
func (c *Compiled) StateNames() []string { return sortedKeys(c.states) }

// ChannelNames returns the declared channel names, sorted.
func (c *Compiled) ChannelNames() []string { return sortedKeys(c.channels) }

// Fault returns the fault by id, or nil.
func (c *Compiled) Fault(id string) *model.Fault {
	for i := range c.Spec.Faults {
		if c.Spec.Faults[i].ID == id {
			return &c.Spec.Faults[i]
		}
	}
	return nil
}

// Effector returns the effector by name, or nil.
func (c *Compiled) Effector(name string) *model.Effector {
	for i := range c.Spec.Effectors {
		if c.Spec.Effectors[i].Name == name {
			return &c.Spec.Effectors[i]
		}
	}
	return nil
}

// Channel returns the channel by name, or nil.
func (c *Compiled) Channel(name string) *model.Channel {
	for i := range c.Spec.Channels {
		if c.Spec.Channels[i].Name == name {
			return &c.Spec.Channels[i]
		}
	}
	return nil
}

// Profile returns the profile by name, or nil.
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
	if err := c.registerSymbols(src); err != nil {
		return nil, err
	}
	if err := crossCheck(c, src); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return c, nil
}
