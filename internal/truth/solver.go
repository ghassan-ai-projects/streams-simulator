// Package truth computes generative ground truth: the three onset
// timestamps, the sealed label records, and the observability solver that
// makes detection latency measured against a real reference rather than a
// guess. Ground truth is reachable only under the director role; the
// operator view contains no reference to it.
package truth

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

// Result is the solved observability of one injection.
type Result struct {
	FirstObservableNS int64    `json:"first_observable_time_ns"`
	UnavoidableNS     int64    `json:"unavoidable_time_ns"`
	EffectiveSigma    float64  `json:"effective_sigma"`
	Observable        bool     `json:"observable"`
	Method            string   `json:"solution_method"`
	Channels          []string `json:"channels"`
}

// SetupCall is a pre-fault effector invocation applied to both the clean
// and faulted worlds, so the scenario context (an aerator running at
// night) exists before the fault lands.
type SetupCall struct {
	Effector  string         `json:"effector"`
	EntityID  string         `json:"entity_id"`
	CommandID string         `json:"command_id"`
	Args      map[string]any `json:"args,omitempty"`
	AtNS      int64          `json:"at_ns"`
}

// Solve computes the onset timestamps for a fault injected at onsetNS on
// entityID. The world ids must be shared so the two runs' substreams align.
func (s *Solver) Solve(entityID, faultID string, onsetNS int64, startNS int64, entityIDs []string, setup []SetupCall) (*Result, error) {
	fault := s.spec.Fault(faultID)
	if fault == nil {
		return nil, fmt.Errorf("truth: unknown fault %q", faultID)
	}
	det := &fault.Observability.Detector
	channels := detectorChannels(det)
	res := &Result{Channels: channels, Method: model.SolveNumeric}

	clean, err := s.buildWorld(entityID, startNS, entityIDs, nil, setup)
	if err != nil {
		return nil, fmt.Errorf("Solve: %w", err)
	}
	faulted, err := s.buildWorld(entityID, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup)
	if err != nil {
		return nil, fmt.Errorf("Solve: %w", err)
	}
	sigma := s.effectiveSigma(entityID, det, entityIDs)

	horizon := onsetNS + s.maxHorizonNS
	if horizon < startNS+s.maxHorizonNS {
		horizon = startNS + s.maxHorizonNS
	}
	t := onsetNS
	if t < startNS {
		t = startNS
	}
	var firstObs, firstUnavoid int64
	for t <= horizon {
		q := s.quantity(entityID, det, clean, faulted, t, entityIDs)
		if firstObs == 0 && q >= fault.Observability.FirstObservableSNR*sigma {
			firstObs = t
		}
		if firstUnavoid == 0 && q >= fault.Observability.UnavoidableSNR*sigma {
			firstUnavoid = t
		}
		if firstUnavoid != 0 {
			break
		}
		t += s.sampleNS
	}
	res.FirstObservableNS = firstObs
	res.UnavoidableNS = firstUnavoid
	res.EffectiveSigma = sigma
	res.Observable = firstObs != 0
	return res, nil
}

func (s *Solver) buildWorld(entityID string, startNS int64, entityIDs []string, faults map[string]int64, setup []SetupCall) (*world.World, error) {
	w, err := world.New(s.spec, s.seed, "w-solver", startNS, world.Options{
		Noiseless:       true,
		EmitDisabled:    true,
		InitialEntities: entityIDs,
		ForceEffectorOK: true,
	})
	if err != nil {
		return nil, fmt.Errorf("buildWorld: %w", err)
	}
	for _, call := range setup {
		if _, err := w.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS); err != nil {
			return nil, fmt.Errorf("truth: setup %s: %w", call.Effector, err)
		}
	}
	for fid, onset := range faults {
		if _, err := w.InjectFault(entityID, fid, onset, nil); err != nil {
			return nil, fmt.Errorf("buildWorld: %w", err)
		}
	}
	return w, nil
}

// quantity evaluates the detector expression: the deviation between the
// faulted and clean worlds at time t.
func (s *Solver) quantity(entityID string, det *model.Detector, clean, faulted *world.World, t int64, entityIDs []string) float64 {
	switch det.Form {
	case model.DetectorSingleChannel:
		f := faulted.Reading(entityID, det.Channel, t)
		c := clean.Reading(entityID, det.Channel, t)
		return math.Abs(f - c)
	case model.DetectorDivergence:
		fa := faulted.Reading(entityID, det.ChannelA, t)
		fb := faulted.Reading(entityID, det.ChannelB, t)
		ca := clean.Reading(entityID, det.ChannelA, t)
		cb := clean.Reading(entityID, det.ChannelB, t)
		return math.Abs((fa - fb) - (ca - cb))
	case model.DetectorPeerResidual:
		// The entity against its siblings: the deviation of E's reading from
		// the mean of the other entities' readings.
		fr := s.peerResidual(entityID, det.Channel, faulted, t)
		cr := s.peerResidual(entityID, det.Channel, clean, t)
		return math.Abs(fr - cr)
	case model.DetectorConservation:
		fi := s.balance(det.Inputs, faulted, t, entityID)
		ci := s.balance(det.Inputs, clean, t, entityID)
		fo := s.balance(det.Outputs, faulted, t, entityID)
		co := s.balance(det.Outputs, clean, t, entityID)
		return math.Abs((fi - fo) - (ci - co))
	}
	return 0
}

func (s *Solver) peerResidual(entityID, channel string, w *world.World, t int64) float64 {
	me := w.Reading(entityID, channel, t)
	var sum float64
	n := 0
	for _, id := range w.EntityIDs() {
		if id == entityID {
			continue
		}
		sum += w.Reading(id, channel, t)
		n++
	}
	if n == 0 {
		return 0
	}
	return me - sum/float64(n)
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
func (s *Solver) effectiveSigma(entityID string, det *model.Detector, entityIDs []string) float64 {
	sig := func(ch string) float64 {
		c := s.spec.Channel(ch)
		if c == nil {
			return 0
		}
		return c.Noise.Sigma
	}
	switch det.Form {
	case model.DetectorSingleChannel:
		return sig(det.Channel)
	case model.DetectorDivergence:
		return math.Hypot(sig(det.ChannelA), sig(det.ChannelB))
	case model.DetectorPeerResidual:
		n := len(entityIDs)
		if n < 2 {
			n = 2
		}
		return sig(det.Channel) * math.Sqrt(1+1/float64(n-1))
	case model.DetectorConservation:
		var sum float64
		for _, c := range append(append([]string{}, det.Inputs...), det.Outputs...) {
			sum += sig(c) * sig(c)
		}
		return math.Sqrt(sum)
	}
	return 0
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
