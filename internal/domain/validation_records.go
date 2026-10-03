package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkChannel(ch *model.Channel) error {
	for _, check := range []func(*model.Channel) error{checks.channelReferences, checks.channelDelivery, checks.channelNoise, checks.channelAvailability} {
		if err := check(ch); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) channelReferences(ch *model.Channel) error {
	c := checks.compiled
	if ch.Observes != "" && !c.HasState(ch.Observes) {
		return checks.bad("channel %q observes undeclared state %q", ch.Name, ch.Observes)
	}
	if ch.ObservationBias != nil && !c.HasState(ch.ObservationBias.State) {
		return checks.bad("channel %q bias references undeclared state %q", ch.Name, ch.ObservationBias.State)
	}
	if ch.Cadence.Mode == "event_driven" && ch.Cadence.TriggerState != "" && !c.HasState(ch.Cadence.TriggerState) {
		return checks.bad("channel %q trigger references undeclared state %q", ch.Name, ch.Cadence.TriggerState)
	}
	return nil
}

func (checks domainChecks) channelDelivery(ch *model.Channel) error {
	if ch.Cadence.Mode == "batch" {
		return checks.bad("channel %q uses batch cadence, which is not implemented", ch.Name)
	}
	if ch.Noise.Model == "pink" {
		return checks.bad("channel %q uses pink noise, which is not implemented", ch.Name)
	}
	return nil
}

func (checks domainChecks) channelNoise(ch *model.Channel) error {
	// Fail closed: declaring no noise with nonzero sigma otherwise uses the gaussian branch.
	if ch.Noise.Model == "none" && ch.Noise.Sigma > 0 {
		return checks.bad("channel %q declares noise model none with sigma %v; sigma must be 0", ch.Name, ch.Noise.Sigma)
	}
	return nil
}

func (checks domainChecks) channelAvailability(ch *model.Channel) error {
	if ch.Availability != nil && ch.Availability.MTTRS <= 0 {
		return checks.bad("channel %q availability requires mttr_s > 0", ch.Name)
	}
	return nil
}

func (checks domainChecks) checkDynamic(d *model.Dynamics) error {
	for _, check := range []func(*model.Dynamics) error{checks.dynamicTarget, checks.dynamicTier, checks.dynamicF1, checks.dynamicF2} {
		if err := check(d); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) dynamicTarget(d *model.Dynamics) error {
	if !checks.compiled.HasState(d.Target) {
		return checks.bad("dynamics target %q is not a declared state", d.Target)
	}
	return nil
}

func (checks domainChecks) dynamicTier(d *model.Dynamics) error {
	// Unsupported tiers fail loudly rather than silently leave simulated state static.
	if d.Tier == "F3" {
		return fmt.Errorf("domain: %s: F3/FMU dynamics are the specified-but-unbuilt seam; not_implemented", checks.source)
	}
	if d.Tier == "F2" {
		return fmt.Errorf("domain: %s: F2 reference models are the specified-but-unbuilt seam; not_implemented", checks.source)
	}
	return nil
}

func (checks domainChecks) dynamicF1(d *model.Dynamics) error {
	if d.F1 != nil {
		if err := checkF1(d.F1, checks.compiled, checks.source); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}

func (checks domainChecks) dynamicF2(d *model.Dynamics) error {
	if d.F2 != nil {
		for name := range d.F2.Inputs {
			if !checks.compiled.HasState(name) {
				return checks.bad("f2 model %q input %q is not a declared state", d.F2.Model, name)
			}
		}
	}
	return nil
}

func (checks domainChecks) checkFault(f *model.Fault) error {
	for _, check := range []func(*model.Fault) error{checks.faultRate, checks.faultStates, checks.faultDetector, checks.faultEffector} {
		if err := check(f); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) faultRate(f *model.Fault) error {
	if f.Onset.RatePerHour < 0 {
		return checks.bad("fault %q onset rate_per_hour must be non-negative", f.ID)
	}
	return nil
}

func (checks domainChecks) faultStates(f *model.Fault) error {
	for _, a := range f.Affects {
		if !checks.compiled.HasState(a.State) {
			return checks.bad("fault %q affects undeclared state %q", f.ID, a.State)
		}
	}
	return nil
}

func (checks domainChecks) faultDetector(f *model.Fault) error {
	if err := checkDetector(&f.Observability.Detector, checks.compiled, checks.source, "fault "+f.ID); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func (checks domainChecks) faultEffector(f *model.Fault) error {
	if f.ExpectedEffector != "" && !checks.compiled.HasEffector(f.ExpectedEffector) {
		return checks.bad("fault %q expects undeclared effector %q", f.ID, f.ExpectedEffector)
	}
	return nil
}

func (checks domainChecks) checkProfile(p *model.Profile) error {
	for name, w := range p.FaultWeights {
		if !checks.compiled.HasFault(name) {
			return checks.bad("profile %q weights undeclared fault %q", p.Name, name)
		}
		if w < 0 {
			return checks.bad("profile %q has negative weight for %q", p.Name, name)
		}
	}
	return nil
}
