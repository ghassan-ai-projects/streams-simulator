package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func newCompiled(spec *model.DomainSpec, digest string) *Compiled {
	return &Compiled{
		Spec: spec, Digest: digest,
		states: map[string]bool{}, channels: map[string]bool{}, faults: map[string]bool{},
		effectors: map[string]bool{}, profiles: map[string]bool{}, channelGain: map[string]float64{},
	}
}

func (c *Compiled) registerSymbols(src string) error {
	for _, register := range []func(string) error{c.registerStates, c.registerChannels, c.registerFaults, c.registerEffectors, c.registerProfiles} {
		if err := register(src); err != nil {
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

func (c *Compiled) registerStates(src string) error {
	for i := range c.Spec.State {
		item := &c.Spec.State[i]
		if err := registerSymbol(c.states, item.Name, "state", itemPointer("state", i)+"/name", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerChannels(src string) error {
	for i := range c.Spec.Channels {
		item := &c.Spec.Channels[i]
		if err := registerSymbol(c.channels, item.Name, "channel", itemPointer("channels", i)+"/name", src); err != nil {
			return err
		}
		c.channelGain[item.Name] = observationGain(item.ObservationGain)
	}
	return nil
}

func (c *Compiled) registerFaults(src string) error {
	for i := range c.Spec.Faults {
		item := &c.Spec.Faults[i]
		if err := registerSymbol(c.faults, item.ID, "fault", itemPointer("faults", i)+"/id", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerEffectors(src string) error {
	for i := range c.Spec.Effectors {
		item := &c.Spec.Effectors[i]
		if err := registerSymbol(c.effectors, item.Name, "effector", itemPointer("effectors", i)+"/name", src); err != nil {
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
