package domain

import (
	"fmt"
	"maps"
	"slices"

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
	if ch.Observes != "" && !checks.hasState(ch.Observes) {
		return checks.bad("/observes", "channel %q observes undeclared state %q", ch.Name, ch.Observes)
	}
	if ch.ObservationBias != nil && !checks.hasState(ch.ObservationBias.State) {
		return checks.bad("/observation_bias/state", "channel %q bias references undeclared state %q", ch.Name, ch.ObservationBias.State)
	}
	if ch.Cadence.Mode == "event_driven" && ch.Cadence.TriggerState != "" && !checks.hasState(ch.Cadence.TriggerState) {
		return checks.bad("/cadence/trigger_state", "channel %q trigger references undeclared state %q", ch.Name, ch.Cadence.TriggerState)
	}
	return nil
}

func (checks domainChecks) channelDelivery(ch *model.Channel) error {
	if ch.Cadence.Mode == "batch" {
		return checks.bad("/cadence/mode", "channel %q uses batch cadence, which is not implemented", ch.Name)
	}
	if ch.Noise.Model == "pink" {
		return checks.bad("/noise/model", "channel %q uses pink noise, which is not implemented", ch.Name)
	}
	return nil
}

func (checks domainChecks) channelNoise(ch *model.Channel) error {
	// Fail closed: declaring no noise with nonzero sigma otherwise uses the gaussian branch.
	if ch.Noise.Model == "none" && ch.Noise.Sigma > 0 {
		return checks.bad("/noise/sigma", "channel %q declares noise model none with sigma %v; sigma must be 0", ch.Name, ch.Noise.Sigma)
	}
	return nil
}

func (checks domainChecks) channelAvailability(ch *model.Channel) error {
	if ch.Availability != nil && ch.Availability.MTTRS <= 0 {
		return checks.bad("/availability/mttr_s", "channel %q availability requires mttr_s > 0", ch.Name)
	}
	return nil
}

func (checks domainChecks) checkDynamic(d *model.Dynamics) error {
	for _, check := range []func(*model.Dynamics) error{checks.dynamicTarget, checks.dynamicTier, checks.dynamicF1, checks.dynamicF2, checks.dynamicSchedule, checks.dynamicPeak, checks.dynamicTripValue} {
		if err := check(d); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) dynamicTarget(d *model.Dynamics) error {
	if !checks.hasState(d.Target) {
		return checks.bad("/target", "dynamics target %q is not a declared state", d.Target)
	}
	return nil
}

func (checks domainChecks) dynamicTier(d *model.Dynamics) error {
	// Unsupported tiers fail loudly rather than silently leave simulated state static.
	if d.Tier == "F3" {
		return checks.bad("/tier", "F3/FMU dynamics are the specified-but-unbuilt seam; not_implemented")
	}
	if d.Tier == "F2" {
		return checks.bad("/tier", "F2 reference models are the specified-but-unbuilt seam; not_implemented")
	}
	return nil
}

func (checks domainChecks) dynamicF1(d *model.Dynamics) error {
	if d.F1 != nil {
		if err := checks.within(checks.at + "/f1").checkF1(d.F1); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}

func (checks domainChecks) dynamicF2(d *model.Dynamics) error {
	if d.F2 != nil {
		for _, name := range slices.Sorted(maps.Keys(d.F2.Inputs)) {
			if !checks.hasState(name) {
				return checks.bad("/f2/inputs"+memberPointer(name), "f2 model %q input %q is not a declared state", d.F2.Model, name)
			}
		}
	}
	return nil
}

func (checks domainChecks) checkFault(f *model.Fault) error {
	for _, check := range []func(*model.Fault) error{checks.faultRate, checks.faultStates, checks.faultDetector, checks.faultEffector, checks.faultDuration} {
		if err := check(f); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) faultRate(f *model.Fault) error {
	if f.Onset.RatePerHour < 0 {
		return checks.bad("/onset/rate_per_hour", "fault %q onset rate_per_hour must be non-negative", f.ID)
	}
	return nil
}

func (checks domainChecks) faultStates(f *model.Fault) error {
	for i, a := range f.Affects {
		if !checks.hasState(a.State) {
			return checks.bad(fmt.Sprintf("/affects/%d/state", i), "fault %q affects undeclared state %q", f.ID, a.State)
		}
	}
	return nil
}

func (checks domainChecks) faultDetector(f *model.Fault) error {
	if err := checks.within(checks.at+"/observability/detector").checkDetector(&f.Observability.Detector, "fault "+f.ID); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func (checks domainChecks) faultEffector(f *model.Fault) error {
	if f.ExpectedEffector != "" && checks.kind.EffectorDecl(f.ExpectedEffector) == nil {
		return checks.bad("/expected_effector", "fault %q expects undeclared effector %q", f.ID, f.ExpectedEffector)
	}
	return nil
}

func (checks domainChecks) checkProfile(p *model.Profile) error {
	for _, name := range slices.Sorted(maps.Keys(p.FaultWeights)) {
		if !checks.compiled.HasFault(name) {
			return checks.bad("/fault_weights"+memberPointer(name), "profile %q weights undeclared fault %q", p.Name, name)
		}
		if p.FaultWeights[name] < 0 {
			return checks.bad("/fault_weights"+memberPointer(name), "profile %q has negative weight for %q", p.Name, name)
		}
	}
	return nil
}
