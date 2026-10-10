// Package domain holds the ground-truth rules: the observability solver that
// finds the onset timestamps by running a clean and a faulted noiseless
// world, the construction of sealed label records, and the sealed store that
// keeps a label from the operator role until its run closes. It performs no
// I/O and reads no clock.
package domain

import (
	"fmt"
	"math"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Solver computes the observability timestamps of a fault by running the
// world twice — once clean, once faulted, both noiseless — and finding when
// the detector quantity crosses the declared SNR thresholds. It reuses the
// verified integrator instead of deriving closed forms per detector form.
type Solver struct {
	spec         *domain.Compiled
	seed         uint64
	sampleNS     int64
	maxHorizonNS int64
}

// NewSolver builds a solver. sampleNS is the resolution of the onset scan;
// maxHorizonNS caps how far the scan looks before declaring a fault
// unobservable.
func NewSolver(spec *domain.Compiled, seed uint64, sampleNS, maxHorizonNS int64) *Solver {
	if sampleNS <= 0 {
		sampleNS = 60 * 1e9
	}
	if maxHorizonNS <= 0 {
		maxHorizonNS = 72 * 3600 * 1e9
	}
	return &Solver{spec: spec, seed: seed, sampleNS: sampleNS, maxHorizonNS: maxHorizonNS}
}

// result is the solved observability of one injection.
type result struct {
	FirstObservableNS int64    `json:"first_observable_time_ns"`
	UnavoidableNS     int64    `json:"unavoidable_time_ns"`
	EffectiveSigma    float64  `json:"effective_sigma"`
	Observable        bool     `json:"observable"`
	Method            string   `json:"solution_method"`
	Channels          []string `json:"channels"`

	// firstFound and unavoidableFound say whether each threshold was
	// crossed: the timestamps alone cannot, because epoch zero is a legal
	// time.
	firstFound, unavoidableFound bool
}

// solve computes the onset timestamps for a fault injected at onsetNS on
// entityID. The world ids must be shared so the two runs' substreams align.
func (s *Solver) solve(entityID, faultID string, onsetNS int64, startNS int64, entityIDs []string, setup []model.SetupCall) (*result, error) {
	fault := s.spec.Fault(faultID)
	if fault == nil {
		return nil, fmt.Errorf("truth: unknown fault %q", faultID)
	}
	if !knownDetectorForm(fault.Observability.Detector.Form) {
		return nil, fmt.Errorf("truth: fault %q declares unknown detector form %q", faultID, fault.Observability.Detector.Form)
	}
	res := &result{Channels: detectorChannels(&fault.Observability.Detector), Method: model.SolveNumeric}
	scan, err := s.prepareScan(entityID, faultID, fault, onsetNS, startNS, entityIDs, setup)
	if err != nil {
		return nil, err
	}
	return scan.searchOnsets(res, fault, onsetNS, startNS), nil
}

func (s *Solver) buildWorld(entityID string, startNS int64, entityIDs []string, faults map[string]int64, setup []model.SetupCall) (*world.World, error) {
	w, err := s.newOracleWorld(startNS, entityIDs)
	if err != nil {
		return nil, err
	}
	if err := applyOracleSetup(w, setup); err != nil {
		return nil, err
	}
	if err := injectOracleFaults(w, entityID, faults); err != nil {
		return nil, err
	}
	return w, nil
}

// quantity evaluates the detector expression: the deviation between the
// faulted and clean worlds at time t.
func (s *Solver) quantity(entityID string, det *model.Detector, clean, faulted *world.World, t int64, entityIDs []string) float64 {
	switch det.Form {
	case model.DetectorSingleChannel:
		return singleChannelDeviation(entityID, det.Channel, clean, faulted, t)
	case model.DetectorDivergence:
		return channelDivergence(entityID, det, clean, faulted, t)
	case model.DetectorPeerResidual:
		return s.peerDeviation(entityID, det.Channel, clean, faulted, t)
	case model.DetectorConservation:
		return s.conservationDeviation(entityID, det, clean, faulted, t)
	}
	return 0
}

func (s *Solver) peerResidual(entityID, channel string, w *world.World, t int64) float64 {
	me := w.Reading(entityID, channel, t)
	sum, count := peerReadings(entityID, channel, w, t)
	if count == 0 {
		return 0
	}
	return me - sum/float64(count)
}

func (s *Solver) balance(channels []string, w *world.World, t int64, entityID string) float64 {
	var sum float64
	for _, c := range channels {
		sum += w.Reading(entityID, c, t)
	}
	return sum
}

// effectiveSigma is the detector's noise floor after the form's propagation
// rule: a divergence combines sigmas in quadrature, a peer residual pools
// the sibling variance, a conservation residual sums over the balance.
// peers is the number of entities the world actually holds: a peer residual
// pools over those, whether or not the caller named them.
func (s *Solver) effectiveSigma(entityID string, det *model.Detector, peers int) float64 {
	switch det.Form {
	case model.DetectorSingleChannel:
		return s.channelSigma(det.Channel)
	case model.DetectorDivergence:
		return math.Hypot(s.channelSigma(det.ChannelA), s.channelSigma(det.ChannelB))
	case model.DetectorPeerResidual:
		return s.peerSigma(det.Channel, peers)
	case model.DetectorConservation:
		return s.balanceSigma(det)
	}
	return 0
}

func knownDetectorForm(form string) bool {
	switch form {
	case model.DetectorSingleChannel, model.DetectorDivergence, model.DetectorPeerResidual, model.DetectorConservation:
		return true
	}
	return false
}

func detectorChannels(det *model.Detector) []string {
	var out []string
	switch det.Form {
	case model.DetectorSingleChannel, model.DetectorPeerResidual:
		out = []string{det.Channel}
	case model.DetectorDivergence:
		out = []string{det.ChannelA, det.ChannelB}
	case model.DetectorConservation:
		out = append(append([]string{}, det.Inputs...), det.Outputs...)
	}
	sort.Strings(out)
	return out
}

func (s *Solver) newOracleWorld(startNS int64, entityIDs []string) (*world.World, error) {
	w, err := world.New(s.spec, s.seed, "w-solver", startNS, world.Options{
		Noiseless:       true,
		EmitDisabled:    true,
		InitialEntities: entityIDs,
		ForceEffectorOK: true,
	})
	if err != nil {
		return nil, fmt.Errorf("buildWorld: %w", err)
	}
	return w, nil
}

func applyOracleSetup(w *world.World, setup []model.SetupCall) error {
	for _, call := range setup {
		if _, err := w.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS); err != nil {
			return fmt.Errorf("truth: setup %s: %w", call.Effector, err)
		}
	}
	return nil
}

func injectOracleFaults(w *world.World, entityID string, faults map[string]int64) error {
	for fid, onset := range faults {
		if _, err := w.InjectFault(entityID, fid, onset, nil); err != nil {
			return fmt.Errorf("buildWorld: %w", err)
		}
	}
	return nil
}
