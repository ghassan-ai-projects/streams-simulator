// Package domain loads, validates and compiles domain specs — the data
// files that define whole simulated worlds. The binary contains no domain
// behavior: every domain in the catalog loads through this one path, and a
// domain that needs a code branch in the binary is a bug in the simulator's
// design, not a missing feature.
package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*Compiled, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("domain: read %s: %w", path, err)
	}
	return Parse(raw, path)
}

// Parse is Load over an in-memory document. src names the source for error
// messages.
func Parse(raw []byte, src string) (*Compiled, error) {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("domain: %s: not valid JSON: %w", src, err)
	}
	sch, err := jsonschema.Compile(mustAny(schemas.DomainSpec()))
	if err != nil {
		return nil, fmt.Errorf("domain: compile contract schema: %w", err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return nil, fmt.Errorf("domain: %s fails domain-spec-v0.1 validation:\n  %s", src, formatErrs(errs))
	}
	var spec model.DomainSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("domain: %s: decode: %w", src, err)
	}
	digest, err := canonical.Digest(doc)
	if err != nil {
		return nil, fmt.Errorf("domain: %s: digest: %w", src, err)
	}
	c, err := compile(&spec, digest, src)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	c.Raw = append([]byte(nil), raw...)
	return c, nil
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// id. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*Compiled, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("domain: list %s: %w", dir, err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	var out []*Compiled
	for _, p := range paths {
		c, err := Load(p)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out = append(out, c)
	}
	return out, nil
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

// FaultNames returns the declared fault ids, sorted.
func (c *Compiled) FaultNames() []string { return sortedKeys(c.faults) }

// EffectorNames returns the declared effector names, sorted.
func (c *Compiled) EffectorNames() []string { return sortedKeys(c.effectors) }

// ProfileNames returns the declared profile names, sorted.
func (c *Compiled) ProfileNames() []string { return sortedKeys(c.profiles) }

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
	c := &Compiled{
		Spec:        spec,
		Digest:      digest,
		states:      map[string]bool{},
		channels:    map[string]bool{},
		faults:      map[string]bool{},
		effectors:   map[string]bool{},
		profiles:    map[string]bool{},
		channelGain: map[string]float64{},
	}
	for i := range spec.State {
		s := &spec.State[i]
		if c.states[s.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate state %q", src, s.Name)
		}
		c.states[s.Name] = true
	}
	for i := range spec.Channels {
		ch := &spec.Channels[i]
		if c.channels[ch.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate channel %q", src, ch.Name)
		}
		c.channels[ch.Name] = true
		gain := ch.ObservationGain
		if gain == 0 {
			gain = 1
		}
		c.channelGain[ch.Name] = gain
	}
	for i := range spec.Faults {
		f := &spec.Faults[i]
		if c.faults[f.ID] {
			return nil, fmt.Errorf("domain: %s: duplicate fault %q", src, f.ID)
		}
		c.faults[f.ID] = true
	}
	for i := range spec.Effectors {
		e := &spec.Effectors[i]
		if c.effectors[e.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate effector %q", src, e.Name)
		}
		c.effectors[e.Name] = true
	}
	for i := range spec.Profiles {
		p := &spec.Profiles[i]
		if c.profiles[p.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate profile %q", src, p.Name)
		}
		c.profiles[p.Name] = true
	}
	if err := crossCheck(c, src); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return c, nil
}
