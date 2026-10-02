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
	c := checks.compiled
	spec := checks.compiled.Spec
	for i := range spec.Channels {
		ch := &spec.Channels[i]
		if ch.Observes != "" && !c.HasState(ch.Observes) {
			return checks.bad("channel %q observes undeclared state %q", ch.Name, ch.Observes)
		}
		if ch.ObservationBias != nil && !c.HasState(ch.ObservationBias.State) {
			return checks.bad("channel %q bias references undeclared state %q", ch.Name, ch.ObservationBias.State)
		}
		if ch.Cadence.Mode == "event_driven" && ch.Cadence.TriggerState != "" && !c.HasState(ch.Cadence.TriggerState) {
			return checks.bad("channel %q trigger references undeclared state %q", ch.Name, ch.Cadence.TriggerState)
		}
		if ch.Cadence.Mode == "batch" {
			return checks.bad("channel %q uses batch cadence, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "pink" {
			return checks.bad("channel %q uses pink noise, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "none" && ch.Noise.Sigma > 0 {
			// Fail closed: a "no noise" declaration with a nonzero sigma
			// would otherwise be silently approximated by the gaussian
			// branch in observe.go.
			return checks.bad("channel %q declares noise model none with sigma %v; sigma must be 0", ch.Name, ch.Noise.Sigma)
		}
		if ch.Availability != nil && ch.Availability.MTTRS <= 0 {
			return checks.bad("channel %q availability requires mttr_s > 0", ch.Name)
		}
	}
	return nil
}

func (checks domainChecks) checkDynamics() error {
	c := checks.compiled
	spec := checks.compiled.Spec
	for i := range spec.Dynamics {
		d := &spec.Dynamics[i]
		if !c.HasState(d.Target) {
			return checks.bad("dynamics target %q is not a declared state", d.Target)
		}
		if d.Tier == "F3" {
			return fmt.Errorf("domain: %s: F3/FMU dynamics are the specified-but-unbuilt seam; not_implemented", checks.source)
		}
		if d.Tier == "F2" {
			// F2 reference models are not built. The integrator must fail
			// loudly rather than silently leave the state static: a domain
			// that used F2 would emit confident wrong values.
			return fmt.Errorf("domain: %s: F2 reference models are the specified-but-unbuilt seam; not_implemented", checks.source)
		}
		if d.F1 != nil {
			if err := checkF1(d.F1, c, checks.source); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
		if d.F2 != nil {
			for name := range d.F2.Inputs {
				if !c.HasState(name) {
					return checks.bad("f2 model %q input %q is not a declared state", d.F2.Model, name)
				}
			}
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
	c := checks.compiled
	spec := checks.compiled.Spec
	for i := range spec.Faults {
		f := &spec.Faults[i]
		if f.Onset.RatePerHour < 0 {
			return checks.bad("fault %q onset rate_per_hour must be non-negative", f.ID)
		}
		for _, a := range f.Affects {
			if !c.HasState(a.State) {
				return checks.bad("fault %q affects undeclared state %q", f.ID, a.State)
			}
		}
		if err := checkDetector(&f.Observability.Detector, c, checks.source, "fault "+f.ID); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if f.ExpectedEffector != "" && !c.HasEffector(f.ExpectedEffector) {
			return checks.bad("fault %q expects undeclared effector %q", f.ID, f.ExpectedEffector)
		}
	}
	return nil
}

func (checks domainChecks) checkEffectors() error {
	c := checks.compiled
	spec := checks.compiled.Spec
	for i := range spec.Effectors {
		e := &spec.Effectors[i]
		for _, d := range e.Effect.StateDeltas {
			if !c.HasState(d.State) {
				return checks.bad("effector %q effect targets undeclared state %q", e.Name, d.State)
			}
		}
		if e.Interlock != nil && !c.HasState(e.Interlock.State) {
			return checks.bad("effector %q interlock references undeclared state %q", e.Name, e.Interlock.State)
		}
		if e.Interlock != nil && e.Interlock.AutonomousAction != nil {
			if !c.HasState(e.Interlock.AutonomousAction.State) {
				return checks.bad("effector %q autonomous action targets undeclared state %q", e.Name, e.Interlock.AutonomousAction.State)
			}
		}
		for _, ch := range e.ConfirmationChannels {
			if !c.HasChannel(ch) {
				return checks.bad("effector %q confirmation channel %q is not declared", e.Name, ch)
			}
		}
		// The dead time is modeled once: an effector that declares its own
		// dead_time_s must not drive a dead_time state, or the delay applies
		// twice.
		if e.Effect.DeadTimeS > 0 {
			for _, d := range e.Effect.StateDeltas {
				if dyn := dynamicsFor(c, d.State); dyn != nil && dyn.F1 != nil && dyn.F1.Form == "dead_time" {
					return checks.bad("effector %q has dead_time_s and drives dead_time state %q: delay would apply twice", e.Name, d.State)
				}
			}
		}
	}
	return nil
}

func (checks domainChecks) checkProfiles() error {
	c := checks.compiled
	spec := checks.compiled.Spec
	for i := range spec.Profiles {
		p := &spec.Profiles[i]
		for name, w := range p.FaultWeights {
			if !c.HasFault(name) {
				return checks.bad("profile %q weights undeclared fault %q", p.Name, name)
			}
			if w < 0 {
				return checks.bad("profile %q has negative weight for %q", p.Name, name)
			}
		}
	}
	return nil
}
