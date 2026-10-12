package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) f1Inputs(f1 *model.F1Dyn) error {
	for i, in := range f1.Inputs {
		if !checks.compiled.HasState(in.State) {
			return checks.bad(fmt.Sprintf("/inputs/%d", i), "f1 form %q input %q is not a declared state", f1.Form, in.State)
		}
	}
	return nil
}

func (checks domainChecks) f1Parameters(f1 *model.F1Dyn) error {
	switch f1.Form {
	case "first_order_lag", "dead_time":
		return checks.f1TimeConstant(f1)
	case "rc_network":
		return checks.f1NetworkConstants(f1)
	case "hysteresis":
		return checks.f1Thresholds(f1)
	case "threshold_integrator":
		return checks.f1Direction(f1)
	case "saturation":
		return checks.f1Clamp(f1)
	}
	return nil
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
	if d.Channel == "" || !checks.compiled.HasChannel(d.Channel) {
		return checks.bad("/channel", "%s: detector %s requires a declared channel", who, d.Form)
	}
	return nil
}

func (checks domainChecks) detectorDivergence(d *model.Detector, who string) error {
	c := checks.compiled
	if d.ChannelA == "" || d.ChannelB == "" || !c.HasChannel(d.ChannelA) || !c.HasChannel(d.ChannelB) {
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
		if !checks.compiled.HasChannel(ch) {
			return checks.bad(fmt.Sprintf("/%s/%d", member, i), "%s: conservation_residual references undeclared channel %q", who, ch)
		}
	}
	return nil
}
