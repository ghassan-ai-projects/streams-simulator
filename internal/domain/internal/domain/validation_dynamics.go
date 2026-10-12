package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) f1Inputs(f1 *model.F1Dyn) error {
	for i, in := range f1.Inputs {
		if !checks.hasState(in.State) {
			return checks.bad(fmt.Sprintf("/inputs/%d", i), "f1 form %q input %q is not a declared state", f1.Form, in.State)
		}
	}
	return nil
}

func (checks domainChecks) f1Parameters(f1 *model.F1Dyn) error {
	if check, known := checks.f1ParameterChecks()[f1.Form]; known {
		return check(f1)
	}
	return nil
}

func (checks domainChecks) f1ParameterChecks() map[string]func(*model.F1Dyn) error {
	return map[string]func(*model.F1Dyn) error{
		"first_order_lag": checks.f1TimeConstant, "dead_time": checks.f1TimeConstant, "rc_network": checks.f1NetworkConstants,
		"hysteresis": checks.f1Thresholds, "threshold_integrator": checks.f1Direction, "saturation": checks.f1Clamp,
		model.FormRegime: checks.f1Selector, model.FormTrip: checks.f1Trip, model.FormAntiRecycle: checks.f1AntiRecycle, model.FormSequencer: checks.f1Sequencer,
	}
}

func (checks domainChecks) f1TimeConstant(f1 *model.F1Dyn) error {
	if f1.TimeConstantS == 0 {
		return checks.bad("/time_constant_s", "f1 form %q requires time_constant_s", f1.Form)
	}
	return nil
}

func (checks domainChecks) f1NetworkConstants(f1 *model.F1Dyn) error {
	if f1.TimeConstantS == 0 || f1.TimeConstant2S == 0 {
		return checks.bad("/time_constant_2_s", "rc_network requires time_constant_s and time_constant_2_s")
	}
	return nil
}

func (checks domainChecks) f1Thresholds(f1 *model.F1Dyn) error {
	if f1.ThresholdLow >= f1.ThresholdHigh {
		return checks.bad("/threshold_low", "hysteresis requires threshold_low < threshold_high")
	}
	return nil
}

func (checks domainChecks) f1Direction(f1 *model.F1Dyn) error {
	if f1.Direction != "below" && f1.Direction != "above" {
		return checks.bad("/direction", "threshold_integrator requires direction below|above")
	}
	return nil
}

func (checks domainChecks) f1Clamp(f1 *model.F1Dyn) error {
	if f1.Clamp == nil || (f1.Clamp.Min == nil && f1.Clamp.Max == nil) {
		return checks.bad("/clamp", "saturation requires a clamp interval")
	}
	return nil
}

func (checks domainChecks) detectorChannel(d *model.Detector, who string) error {
	if d.Channel == "" || !checks.hasChannel(d.Channel) {
		return checks.bad("/channel", "%s: detector %s requires a declared channel", who, d.Form)
	}
	return nil
}

func (checks domainChecks) detectorDivergence(d *model.Detector, who string) error {
	if d.ChannelA == "" || d.ChannelB == "" || !checks.hasChannel(d.ChannelA) || !checks.hasChannel(d.ChannelB) {
		return checks.bad("/channel_a", "%s: channel_divergence requires declared channel_a and channel_b", who)
	}
	return nil
}

func (checks domainChecks) detectorConservation(d *model.Detector, who string) error {
	if err := checks.undeclaredChannels("inputs", d.Inputs, who); err != nil {
		return err
	}
	return checks.undeclaredChannels("outputs", d.Outputs, who)
}

func (checks domainChecks) undeclaredChannels(member string, names []string, who string) error {
	for i, ch := range names {
		if !checks.hasChannel(ch) {
			return checks.bad(fmt.Sprintf("/%s/%d", member, i), "%s: conservation_residual references undeclared channel %q", who, ch)
		}
	}
	return nil
}

func (checks domainChecks) hasState(name string) bool { return checks.kind.StateDecl(name) != nil }

func (checks domainChecks) hasChannel(name string) bool { return checks.kind.ChannelDecl(name) != nil }
