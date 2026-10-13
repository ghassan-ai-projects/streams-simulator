package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func crossCheck(c *Compiled, src string) error {
	checks := domainChecks{compiled: c, source: src}
	if err := checks.checkInstances(); err != nil {
		return err
	}
	for i := range c.Spec.Kinds {
		if err := checks.inKind(i).checkKind(); err != nil {
			return err
		}
	}
	return checks.checkRest()
}

func (checks domainChecks) checkRest() error {
	for _, check := range []func() error{checks.checkOverrides, checks.checkStructure, checks.checkProfiles} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

type domainChecks struct {
	compiled *Compiled
	kind     *model.Kind
	subject  *model.Kind
	source   string
	at       string
}

func (checks domainChecks) inKind(index int) domainChecks {
	checks.kind = &checks.compiled.Spec.Kinds[index]
	checks.subject = checks.kind
	checks.at = itemPointer("kinds", index)
	return checks
}

func (checks domainChecks) checkKind() error {
	for _, check := range []func() error{checks.checkChannels, checks.checkDynamics, checks.checkDynamicsGraph, checks.checkFlows, checks.checkStoreSteps, checks.checkStability, checks.checkFaults, checks.checkEffectors, checks.checkBoundaries, checks.checkKPIs} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) within(at string) domainChecks {
	checks.at = at
	return checks
}

func (checks domainChecks) bad(member, format string, args ...any) error {
	return invalidDomain(checks.source, checks.at+member, fmt.Sprintf(format, args...))
}

func invalidDomain(src, pointer, message string) *model.ValidationError {
	return &model.ValidationError{Contract: model.ContractDomainSpec, Source: src,
		Summary: fmt.Sprintf("domain: %s: %s", src, message), Issues: []model.Issue{model.LocatedIssue(pointer, message)}}
}

func (checks domainChecks) each(section string, count int, check func(domainChecks, int) error) error {
	for i := 0; i < count; i++ {
		if err := check(checks.within(checks.at+itemPointer(section, i)), i); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkChannels() error {
	kind := checks.subject
	return checks.each("channels", len(kind.Channels), func(at domainChecks, i int) error { return at.checkChannel(&kind.Channels[i]) })
}

func (checks domainChecks) checkDynamics() error {
	kind := checks.subject
	return checks.each("dynamics", len(kind.Dynamics), func(at domainChecks, i int) error { return at.checkDynamic(&kind.Dynamics[i]) })
}

func (checks domainChecks) checkDynamicsGraph() error {
	return rejectDynamicsCycles(checks.subject, checks.at, checks.source)
}

func (checks domainChecks) checkFaults() error {
	kind := checks.subject
	return checks.each("faults", len(kind.Faults), func(at domainChecks, i int) error { return at.checkFault(&kind.Faults[i]) })
}

func (checks domainChecks) checkEffectors() error {
	kind := checks.subject
	return checks.each("effectors", len(kind.Effectors), func(at domainChecks, i int) error { return at.checkEffector(&kind.Effectors[i]) })
}

func (checks domainChecks) checkProfiles() error {
	spec := checks.compiled.Spec
	return checks.each("profiles", len(spec.Profiles), func(at domainChecks, i int) error { return at.checkProfile(&spec.Profiles[i]) })
}
