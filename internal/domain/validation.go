package domain

import (
	"fmt"
)

// crossCheck validates every cross-reference in the spec: channels observe
// declared states, faults affect declared states, effectors touch declared
// states, detector expressions name declared channels, and so on. The JSON
// schema cannot express these.
func crossCheck(c *Compiled, src string) error {
	spec := c.Spec
	bad := func(format string, args ...any) error {
		return fmt.Errorf("domain: %s: %s", src, fmt.Sprintf(format, args...))
	}
	for i := range spec.Channels {
		ch := &spec.Channels[i]
		if ch.Observes != "" && !c.HasState(ch.Observes) {
			return bad("channel %q observes undeclared state %q", ch.Name, ch.Observes)
		}
		if ch.ObservationBias != nil && !c.HasState(ch.ObservationBias.State) {
			return bad("channel %q bias references undeclared state %q", ch.Name, ch.ObservationBias.State)
		}
		if ch.Cadence.Mode == "event_driven" && ch.Cadence.TriggerState != "" && !c.HasState(ch.Cadence.TriggerState) {
			return bad("channel %q trigger references undeclared state %q", ch.Name, ch.Cadence.TriggerState)
		}
		if ch.Cadence.Mode == "batch" {
			return bad("channel %q uses batch cadence, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "pink" {
			return bad("channel %q uses pink noise, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "none" && ch.Noise.Sigma > 0 {
			// Fail closed: a "no noise" declaration with a nonzero sigma
			// would otherwise be silently approximated by the gaussian
			// branch in observe.go.
			return bad("channel %q declares noise model none with sigma %v; sigma must be 0", ch.Name, ch.Noise.Sigma)
		}
		if ch.Availability != nil && ch.Availability.MTTRS <= 0 {
			return bad("channel %q availability requires mttr_s > 0", ch.Name)
		}
	}
	for i := range spec.Dynamics {
		d := &spec.Dynamics[i]
		if !c.HasState(d.Target) {
			return bad("dynamics target %q is not a declared state", d.Target)
		}
		if d.Tier == "F3" {
			return fmt.Errorf("domain: %s: F3/FMU dynamics are the specified-but-unbuilt seam; not_implemented", src)
		}
		if d.Tier == "F2" {
			// F2 reference models are not built. The integrator must fail
			// loudly rather than silently leave the state static: a domain
			// that used F2 would emit confident wrong values.
			return fmt.Errorf("domain: %s: F2 reference models are the specified-but-unbuilt seam; not_implemented", src)
		}
		if d.F1 != nil {
			if err := checkF1(d.F1, c, src); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
		if d.F2 != nil {
			for name := range d.F2.Inputs {
				if !c.HasState(name) {
					return bad("f2 model %q input %q is not a declared state", d.F2.Model, name)
				}
			}
		}
	}
	if err := rejectDynamicsCycles(spec, c, src); err != nil {
		return err
	}
	for i := range spec.Faults {
		f := &spec.Faults[i]
		if f.Onset.RatePerHour < 0 {
			return bad("fault %q onset rate_per_hour must be non-negative", f.ID)
		}
		for _, a := range f.Affects {
			if !c.HasState(a.State) {
				return bad("fault %q affects undeclared state %q", f.ID, a.State)
			}
		}
		if err := checkDetector(&f.Observability.Detector, c, src, "fault "+f.ID); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if f.ExpectedEffector != "" && !c.HasEffector(f.ExpectedEffector) {
			return bad("fault %q expects undeclared effector %q", f.ID, f.ExpectedEffector)
		}
	}
	for i := range spec.Effectors {
		e := &spec.Effectors[i]
		for _, d := range e.Effect.StateDeltas {
			if !c.HasState(d.State) {
				return bad("effector %q effect targets undeclared state %q", e.Name, d.State)
			}
		}
		if e.Interlock != nil && !c.HasState(e.Interlock.State) {
			return bad("effector %q interlock references undeclared state %q", e.Name, e.Interlock.State)
		}
		if e.Interlock != nil && e.Interlock.AutonomousAction != nil {
			if !c.HasState(e.Interlock.AutonomousAction.State) {
				return bad("effector %q autonomous action targets undeclared state %q", e.Name, e.Interlock.AutonomousAction.State)
			}
		}
		for _, ch := range e.ConfirmationChannels {
			if !c.HasChannel(ch) {
				return bad("effector %q confirmation channel %q is not declared", e.Name, ch)
			}
		}
		// The dead time is modeled once: an effector that declares its own
		// dead_time_s must not drive a dead_time state, or the delay applies
		// twice.
		if e.Effect.DeadTimeS > 0 {
			for _, d := range e.Effect.StateDeltas {
				if dyn := dynamicsFor(c, d.State); dyn != nil && dyn.F1 != nil && dyn.F1.Form == "dead_time" {
					return bad("effector %q has dead_time_s and drives dead_time state %q: delay would apply twice", e.Name, d.State)
				}
			}
		}
	}
	for i := range spec.Profiles {
		p := &spec.Profiles[i]
		for name, w := range p.FaultWeights {
			if !c.HasFault(name) {
				return bad("profile %q weights undeclared fault %q", p.Name, name)
			}
			if w < 0 {
				return bad("profile %q has negative weight for %q", p.Name, name)
			}
		}
	}
	return nil
}
