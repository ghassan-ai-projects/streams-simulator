package domain

import (
	"fmt"
)

// crossCheck validates every cross-reference in the spec: channels observe
// declared states, faults affect declared states, effectors touch declared
// states, detector expressions name declared channels, and so on. The JSON
// schema cannot express these.
func crossCheck(c *Compiled, src string) error {
	checks := domainChecks{compiled: c, source: src}
	for _, check := range []func() error{checks.checkChannels, checks.checkDynamics, checks.checkDynamicsGraph, checks.checkFaults, checks.checkEffectors, checks.checkProfiles} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

type domainChecks struct {
	compiled *Compiled
	source   string
}

func (checks domainChecks) bad(format string, args ...any) error {
	return fmt.Errorf("domain: %s: %s", checks.source, fmt.Sprintf(format, args...))
}

func (checks domainChecks) checkChannels() error {
	for i := range checks.compiled.Spec.Channels {
		if err := checks.checkChannel(&checks.compiled.Spec.Channels[i]); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkDynamics() error {
	for i := range checks.compiled.Spec.Dynamics {
		if err := checks.checkDynamic(&checks.compiled.Spec.Dynamics[i]); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkDynamicsGraph() error {
	if err := rejectDynamicsCycles(checks.compiled.Spec, checks.compiled, checks.source); err != nil {
		return err
	}
	return nil
}

func (checks domainChecks) checkFaults() error {
	for i := range checks.compiled.Spec.Faults {
		if err := checks.checkFault(&checks.compiled.Spec.Faults[i]); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkEffectors() error {
	for i := range checks.compiled.Spec.Effectors {
		if err := checks.checkEffector(&checks.compiled.Spec.Effectors[i]); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkProfiles() error {
	for i := range checks.compiled.Spec.Profiles {
		if err := checks.checkProfile(&checks.compiled.Spec.Profiles[i]); err != nil {
			return err
		}
	}
	return nil
}
