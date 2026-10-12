package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (c *Compiled) registerOverrides(src string) error {
	for i := range c.Spec.Instances {
		set := &c.Spec.Instances[i]
		if set.Override == nil {
			continue
		}
		if err := c.registerOverride(set, itemPointer("instances", i)+"/override", src); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) registerOverride(set *model.InstanceSet, at, src string) error {
	added := addedKind(set)
	if err := registerStateList(stateNameSet(c.Kind(set.Kind)), added.State, at, src); err != nil {
		return err
	}
	for _, register := range []func(*model.Kind, string, string) error{c.registerChannels, c.registerFaults, c.registerEffectors} {
		if err := register(added, at, src); err != nil {
			return err
		}
	}
	return nil
}

func stateNameSet(kind *model.Kind) map[string]bool {
	names := map[string]bool{}
	for i := range kind.State {
		names[kind.State[i].Name] = true
	}
	return names
}

func addedKind(set *model.InstanceSet) *model.Kind {
	ov := set.Override
	return &model.Kind{Name: set.Kind, State: ov.State, Dynamics: ov.Dynamics, Channels: ov.Channels, Faults: ov.Faults, Effectors: ov.Effectors}
}

func (checks domainChecks) checkOverrides() error {
	return checks.each("instances", len(checks.compiled.Spec.Instances), func(at domainChecks, i int) error {
		set := &checks.compiled.Spec.Instances[i]
		if set.Override == nil {
			return nil
		}
		return at.within(at.at+"/override").checkOverride(i, set)
	})
}

func (checks domainChecks) checkOverride(index int, set *model.InstanceSet) error {
	base := checks.compiled.Kind(set.Kind)
	if err := checks.checkNarrowing(base, set.Override.Narrow); err != nil {
		return err
	}
	effective := effectiveKind(base, set.Override)
	checks.compiled.effective[index] = effective
	return checks.checkAdditions(effective, set)
}

func (checks domainChecks) checkAdditions(effective *model.Kind, set *model.InstanceSet) error {
	checks.kind, checks.subject = effective, addedKind(set)
	if err := checks.addedDynamicsTargets(); err != nil {
		return err
	}
	return checks.checkKind()
}

func (checks domainChecks) addedDynamicsTargets() error {
	added := stateNameSet(&model.Kind{State: checks.subject.State})
	return checks.each("dynamics", len(checks.subject.Dynamics), func(at domainChecks, i int) error {
		target := checks.subject.Dynamics[i].Target
		if !added[target] {
			return at.bad("/target", "override dynamics may drive only states the override adds, not %q", target)
		}
		return nil
	})
}

func effectiveKind(base *model.Kind, ov *model.Override) *model.Kind {
	effective := *base
	effective.State = append(slices.Clone(base.State), ov.State...)
	effective.Dynamics = append(slices.Clone(base.Dynamics), ov.Dynamics...)
	effective.Channels = append(slices.Clone(base.Channels), ov.Channels...)
	effective.Faults = append(slices.Clone(base.Faults), ov.Faults...)
	effective.Effectors = append(slices.Clone(base.Effectors), ov.Effectors...)
	if ov.Narrow != nil {
		applyNarrowing(&effective, ov.Narrow)
	}
	return &effective
}

func (c *Compiled) indexLookup() {
	c.lookup = nil
	for i := range c.Spec.Kinds {
		c.lookup = append(c.lookup, &c.Spec.Kinds[i])
	}
	for i := range c.Spec.Instances {
		if c.Spec.Instances[i].Override != nil {
			c.lookup = append(c.lookup, c.effective[i])
		}
	}
}

func (c *Compiled) DeclaredChannels() []*model.Channel {
	return declaredOnce(c.lookup, func(k *model.Kind) []model.Channel { return k.Channels }, func(ch *model.Channel) string { return ch.Name })
}

func (c *Compiled) DeclaredEffectors() []*model.Effector {
	return declaredOnce(c.lookup, func(k *model.Kind) []model.Effector { return k.Effectors }, func(e *model.Effector) string { return e.Name })
}

func declaredOnce[T any](kinds []*model.Kind, items func(*model.Kind) []T, name func(*T) string) []*T {
	var out []*T
	seen := map[string]bool{}
	for _, kind := range kinds {
		list := items(kind)
		for i := range list {
			if !seen[name(&list[i])] {
				seen[name(&list[i])] = true
				out = append(out, &list[i])
			}
		}
	}
	return out
}

func (c *Compiled) EntityChannel(entityID, name string) *model.Channel {
	if kind := c.InstanceKind(c.InstanceSetFor(entityID)); kind != nil {
		if ch := kind.ChannelDecl(name); ch != nil {
			return ch
		}
	}
	return c.Channel(name)
}
