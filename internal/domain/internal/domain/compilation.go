package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func newCompiled(spec *model.DomainSpec, digest string) *Compiled {
	return &Compiled{
		Spec: spec, Digest: digest,
		channels: map[string]bool{}, faults: map[string]bool{},
		effectors: map[string]bool{}, profiles: map[string]bool{}, channelGain: map[string]float64{},
	}
}

func (c *Compiled) registerSymbols(src string) error {
	if err := c.registerKinds(src); err != nil {
		return err
	}
	return c.registerProfiles(src)
}

func (c *Compiled) registerKinds(src string) error {
	names := map[string]bool{}
	for i := range c.Spec.Kinds {
		kind := &c.Spec.Kinds[i]
		at := itemPointer("kinds", i)
		if err := registerSymbol(names, kind.Name, "kind", at+"/name", src); err != nil {
			return err
		}
		if err := c.registerKind(kind, at, src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerKind(kind *model.Kind, at, src string) error {
	for _, register := range []func(*model.Kind, string, string) error{c.registerStates, c.registerChannels, c.registerFaults, c.registerEffectors} {
		if err := register(kind, at, src); err != nil {
			return err
		}
	}
	return nil
}

func registerSymbol(symbols map[string]bool, name, kind, pointer, src string) error {
	if symbols[name] {
		return invalidDomain(src, pointer, fmt.Sprintf("duplicate %s %q", kind, name))
	}
	symbols[name] = true
	return nil
}

func (c *Compiled) registerStates(kind *model.Kind, at, src string) error {
	states := map[string]bool{}
	for i := range kind.State {
		item := &kind.State[i]
		if err := registerSymbol(states, item.Name, "state", at+itemPointer("state", i)+"/name", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerChannels(kind *model.Kind, at, src string) error {
	for i := range kind.Channels {
		item := &kind.Channels[i]
		if err := registerSymbol(c.channels, item.Name, "channel", at+itemPointer("channels", i)+"/name", src); err != nil {
			return err
		}
		c.channelGain[item.Name] = observationGain(item.ObservationGain)
	}
	return nil
}

func (c *Compiled) registerFaults(kind *model.Kind, at, src string) error {
	for i := range kind.Faults {
		item := &kind.Faults[i]
		if err := registerSymbol(c.faults, item.ID, "fault", at+itemPointer("faults", i)+"/id", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerEffectors(kind *model.Kind, at, src string) error {
	for i := range kind.Effectors {
		item := &kind.Effectors[i]
		if err := registerSymbol(c.effectors, item.Name, "effector", at+itemPointer("effectors", i)+"/name", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerProfiles(src string) error {
	for i := range c.Spec.Profiles {
		item := &c.Spec.Profiles[i]
		if err := registerSymbol(c.profiles, item.Name, "profile", itemPointer("profiles", i)+"/name", src); err != nil {
			return err
		}
	}
	return nil
}

func observationGain(gain float64) float64 {
	if gain == 0 {
		return 1
	}
	return gain
}
