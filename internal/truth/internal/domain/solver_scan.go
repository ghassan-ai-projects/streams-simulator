package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type observabilityScan struct {
	solver         *Solver
	entity         string
	fault          *model.Fault
	detector       *model.Detector
	clean, faulted *world.World
	entityIDs      []string
	sigma          float64
	res            *result
	recorded       bool
}

func (s *Solver) prepareScan(entity, faultID string, fault *model.Fault, onsetNS, startNS int64, entityIDs []string, setup []model.SetupCall) (*observabilityScan, error) {
	clean, err := s.buildWorld(entity, startNS, entityIDs, nil, setup)
	if err != nil {
		return nil, fmt.Errorf("solve: %w", err)
	}
	pulse := fault.Onset.Shape == model.ShapePulse
	faulted, err := s.buildOracle(entity, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup, pulse)
	if err != nil {
		return nil, fmt.Errorf("solve: %w", err)
	}
	return newObservabilityScan(s, entity, fault, clean, faulted, entityIDs).judgedAtRecordingsWhen(pulse, onsetNS), nil
}

func newObservabilityScan(s *Solver, entity string, fault *model.Fault, clean, faulted *world.World, entityIDs []string) *observabilityScan {
	detector := &fault.Observability.Detector
	return &observabilityScan{
		solver: s, entity: entity, fault: fault, detector: detector,
		clean: clean, faulted: faulted, entityIDs: entityIDs,
		sigma: s.effectiveSigma(entity, detector, len(clean.PeerIDs(entity))),
		res:   &result{Channels: detectorChannels(detector), Method: model.SolveNumeric},
	}
}

func (scan *observabilityScan) searchOnsets(onsetNS, startNS int64) (*result, error) {
	horizon := max(onsetNS+scan.solver.maxHorizonNS, startNS+scan.solver.maxHorizonNS)
	for at := max(onsetNS, startNS); at <= horizon && !scan.res.unavoidableFound; at += scan.solver.sampleNS {
		if err := scan.advanceTo(at); err != nil {
			return nil, err
		}
		if !scan.recorded {
			scan.observeThresholds(at)
		}
	}
	scan.res.EffectiveSigma = scan.sigma
	scan.res.Observable = scan.res.firstFound
	return scan.res, nil
}

func (scan *observabilityScan) advanceTo(at int64) error {
	for _, w := range []*world.World{scan.faulted, scan.clean} {
		if _, _, err := w.Advance(at); err != nil {
			return fmt.Errorf("truth: advance the oracle worlds to %d: %w", at, err)
		}
	}
	return nil
}

func (scan *observabilityScan) observeThresholds(at int64) {
	res, observability := scan.res, scan.fault.Observability
	quantity := scan.solver.quantity(scan.entity, scan.detector, scan.clean, scan.faulted, at, scan.entityIDs)
	if !res.firstFound && quantity > 0 && quantity >= observability.FirstObservableSNR*scan.sigma {
		res.FirstObservableNS, res.firstFound = at, true
	}
	if !res.unavoidableFound && quantity > 0 && quantity >= observability.UnavoidableSNR*scan.sigma {
		res.UnavoidableNS, res.unavoidableFound = at, true
	}
}
