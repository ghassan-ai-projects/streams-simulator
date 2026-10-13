package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (checks domainChecks) faultObservation(f *model.Fault) error {
	o := f.Observation
	if o == nil {
		return nil
	}
	ch := checks.kind.ChannelDecl(o.Channel)
	if ch == nil {
		return checks.bad("/observation/channel", "fault %q observes undeclared channel %q", f.ID, o.Channel)
	}
	if err := checks.observedChannel(o, ch); err != nil {
		return err
	}
	return checks.observationDetector(f)
}

func (checks domainChecks) observationDetector(f *model.Fault) error {
	if d := f.Observability.Detector; d.Form != model.DetectorSingleChannel || d.Channel != f.Observation.Channel {
		return checks.bad("/observability/detector", "an observation fault is detected on its own channel %q by single_channel_snr", f.Observation.Channel)
	}
	return nil
}

func (checks domainChecks) observedChannel(o *model.ObservationFault, ch *model.Channel) error {
	if o.Form != model.ObservationDropout && ch.ValueType != "number" && ch.ValueType != "counter" {
		return checks.bad("/observation/form", "a %s fault needs a number or counter channel; %q is %s", o.Form, o.Channel, ch.ValueType)
	}
	if o.Form == model.ObservationNoise && ch.Noise.Sigma <= 0 {
		return checks.bad("/observation/form", "a noise fault scales the channel's noise, and %q has none", o.Channel)
	}
	return nil
}
