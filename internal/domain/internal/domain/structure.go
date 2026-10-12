package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkStructure() error {
	for _, check := range []func() error{checks.checkSystemKinds, checks.checkLinkTypes, checks.checkLinks, checks.checkBoards} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkSystemKinds() error {
	spec := checks.compiled.Spec
	return checks.each("kinds", len(spec.Kinds), func(at domainChecks, i int) error {
		if !spec.Kinds[i].System {
			return nil
		}
		if err := at.systemSet(&spec.Kinds[i]); err != nil {
			return err
		}
		return at.systemPeers(&spec.Kinds[i])
	})
}

func (checks domainChecks) systemSet(kind *model.Kind) error {
	sets := checks.setsOf(kind.Name)
	if len(sets) != 1 {
		return checks.bad("/system", "system kind %q needs exactly one instance set, has %d", kind.Name, len(sets))
	}
	set := &checks.compiled.Spec.Instances[sets[0]]
	if len(set.Names) != 1 || set.IDTemplate != "" {
		return checks.within(itemPointer("instances", sets[0])).bad("/names", "system kind %q needs one named instance and no id_template", kind.Name)
	}
	return nil
}

func (checks domainChecks) setsOf(kind string) []int {
	var sets []int
	for i := range checks.compiled.Spec.Instances {
		if checks.compiled.Spec.Instances[i].Kind == kind {
			sets = append(sets, i)
		}
	}
	return sets
}

func (checks domainChecks) systemPeers(kind *model.Kind) error {
	return checks.each("faults", len(kind.Faults), func(at domainChecks, i int) error {
		if kind.Faults[i].Observability.Detector.Form == model.DetectorPeerResidual {
			return at.bad("/observability/detector/form", "system kind %q has one entity, so fault %q has no peers", kind.Name, kind.Faults[i].ID)
		}
		return nil
	})
}

func (checks domainChecks) checkLinkTypes() error {
	spec := checks.compiled.Spec
	names := map[string]bool{}
	return checks.each("link_types", len(spec.LinkTypes), func(at domainChecks, i int) error {
		item := &spec.LinkTypes[i]
		if err := registerSymbol(names, item.Name, "link type", at.at+"/name", checks.source); err != nil {
			return err
		}
		return at.linkTypeKinds(item)
	})
}

func (checks domainChecks) linkTypeKinds(item *model.LinkType) error {
	if checks.compiled.Kind(item.From) == nil {
		return checks.bad("/from", "link type %q names kind %q, which is not declared", item.Name, item.From)
	}
	if checks.compiled.Kind(item.To) == nil {
		return checks.bad("/to", "link type %q names kind %q, which is not declared", item.Name, item.To)
	}
	return nil
}
