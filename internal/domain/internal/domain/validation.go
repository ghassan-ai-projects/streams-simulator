package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// crossCheck validates every cross-reference in the spec: channels observe
// declared states, faults affect declared states, effectors touch declared
// states, detector expressions name declared channels, and so on. The JSON
// schema cannot express these.
func crossCheck(c *Compiled, src string) error {
	checks := domainChecks{compiled: c, source: src}
	for _, check := range []func() error{checks.checkInstances, checks.checkChannels, checks.checkDynamics, checks.checkDynamicsGraph, checks.checkFaults, checks.checkEffectors, checks.checkProfiles} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

type domainChecks struct {
	compiled *Compiled
	source   string
	at       string
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
		Summary: fmt.Sprintf("domain: %s: %s", src, message), Issues: []model.Issue{model.LocatedIssue(documentPointer(pointer), message)}}
}

func (checks domainChecks) each(section string, count int, check func(domainChecks, int) error) error {
	for i := 0; i < count; i++ {
		if err := check(checks.within(itemPointer(section, i)), i); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkChannels() error {
	spec := checks.compiled.Spec
	return checks.each("channels", len(spec.Channels), func(at domainChecks, i int) error { return at.checkChannel(&spec.Channels[i]) })
}

func (checks domainChecks) checkDynamics() error {
	spec := checks.compiled.Spec
	return checks.each("dynamics", len(spec.Dynamics), func(at domainChecks, i int) error { return at.checkDynamic(&spec.Dynamics[i]) })
}

func (checks domainChecks) checkDynamicsGraph() error {
	return rejectDynamicsCycles(checks.compiled.Spec, checks.compiled, checks.source)
}

func (checks domainChecks) checkFaults() error {
	spec := checks.compiled.Spec
	return checks.each("faults", len(spec.Faults), func(at domainChecks, i int) error { return at.checkFault(&spec.Faults[i]) })
}

func (checks domainChecks) checkEffectors() error {
	spec := checks.compiled.Spec
	return checks.each("effectors", len(spec.Effectors), func(at domainChecks, i int) error { return at.checkEffector(&spec.Effectors[i]) })
}

func (checks domainChecks) checkProfiles() error {
	spec := checks.compiled.Spec
	return checks.each("profiles", len(spec.Profiles), func(at domainChecks, i int) error { return at.checkProfile(&spec.Profiles[i]) })
}
