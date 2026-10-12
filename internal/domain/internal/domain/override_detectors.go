package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (checks domainChecks) addedPeerResiduals() error {
	added := checks.subject
	return checks.each("faults", len(added.Faults), func(at domainChecks, i int) error {
		detector := &added.Faults[i].Observability.Detector
		if detector.Form == model.DetectorPeerResidual && added.ChannelDecl(detector.Channel) != nil {
			return at.bad("/observability/detector/channel", "peer_residual on channel %q, which only this override adds, has no peers to compare with", detector.Channel)
		}
		return nil
	})
}
