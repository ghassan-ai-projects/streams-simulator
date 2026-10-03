package truth

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func singleChannelDeviation(entity, channel string, clean, faulted *world.World, at int64) float64 {
	faulty := faulted.Reading(entity, channel, at)
	baseline := clean.Reading(entity, channel, at)
	return math.Abs(faulty - baseline)
}

func channelDivergence(entity string, det *model.Detector, clean, faulted *world.World, at int64) float64 {
	fa := faulted.Reading(entity, det.ChannelA, at)
	fb := faulted.Reading(entity, det.ChannelB, at)
	ca := clean.Reading(entity, det.ChannelA, at)
	cb := clean.Reading(entity, det.ChannelB, at)
	return math.Abs((fa - fb) - (ca - cb))
}

func (s *Solver) peerDeviation(entity, channel string, clean, faulted *world.World, at int64) float64 {
	// Compare the entity's deviation from the mean of its siblings in each world.
	faulty := s.peerResidual(entity, channel, faulted, at)
	baseline := s.peerResidual(entity, channel, clean, at)
	return math.Abs(faulty - baseline)
}

func (s *Solver) conservationDeviation(entity string, det *model.Detector, clean, faulted *world.World, at int64) float64 {
	fi := s.balance(det.Inputs, faulted, at, entity)
	ci := s.balance(det.Inputs, clean, at, entity)
	fo := s.balance(det.Outputs, faulted, at, entity)
	co := s.balance(det.Outputs, clean, at, entity)
	return math.Abs((fi - fo) - (ci - co))
}

func peerReadings(entity, channel string, w *world.World, at int64) (float64, int) {
	var sum float64
	count := 0
	for _, id := range w.EntityIDs() {
		if id == entity {
			continue
		}
		sum += w.Reading(id, channel, at)
		count++
	}
	return sum, count
}

func (s *Solver) channelSigma(channel string) float64 {
	declaration := s.spec.Channel(channel)
	if declaration == nil {
		return 0
	}
	return declaration.Noise.Sigma
}

func (s *Solver) peerSigma(channel string, count int) float64 {
	count = max(count, 2)
	return s.channelSigma(channel) * math.Sqrt(1+1/float64(count-1))
}

func (s *Solver) balanceSigma(det *model.Detector) float64 {
	var variance float64
	for _, channel := range append(append([]string{}, det.Inputs...), det.Outputs...) {
		variance += s.channelSigma(channel) * s.channelSigma(channel)
	}
	return math.Sqrt(variance)
}
