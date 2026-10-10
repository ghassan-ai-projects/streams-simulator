package truth

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type observabilityScan struct {
	solver         *Solver
	entity         string
	detector       *model.Detector
	clean, faulted *world.World
	entityIDs      []string
	sigma          float64
}

func (s *Solver) prepareScan(entity, faultID string, fault *model.Fault, onsetNS, startNS int64, entityIDs []string, setup []model.SetupCall) (*observabilityScan, error) {
	clean, err := s.buildWorld(entity, startNS, entityIDs, nil, setup)
	if err != nil {
		return nil, fmt.Errorf("solve: %w", err)
	}
	faulted, err := s.buildWorld(entity, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup)
	if err != nil {
		return nil, fmt.Errorf("solve: %w", err)
	}
	return newObservabilityScan(s, entity, fault, clean, faulted, entityIDs), nil
}

func newObservabilityScan(s *Solver, entity string, fault *model.Fault, clean, faulted *world.World, entityIDs []string) *observabilityScan {
	detector := &fault.Observability.Detector
	return &observabilityScan{
		solver: s, entity: entity, detector: detector,
		clean: clean, faulted: faulted, entityIDs: entityIDs,
		sigma: s.effectiveSigma(entity, detector, entityIDs),
	}
}

func (scan *observabilityScan) searchOnsets(res *result, fault *model.Fault, onsetNS, startNS int64) *result {
	horizon := max(onsetNS+scan.solver.maxHorizonNS, startNS+scan.solver.maxHorizonNS)
	for at := max(onsetNS, startNS); at <= horizon; at += scan.solver.sampleNS {
		if scan.observeThresholds(res, fault, at) {
			break
		}
	}
	res.EffectiveSigma = scan.sigma
	res.Observable = res.FirstObservableNS != 0
	return res
}

func (scan *observabilityScan) observeThresholds(res *result, fault *model.Fault, at int64) bool {
	quantity := scan.solver.quantity(scan.entity, scan.detector, scan.clean, scan.faulted, at, scan.entityIDs)
	if res.FirstObservableNS == 0 && quantity >= fault.Observability.FirstObservableSNR*scan.sigma {
		res.FirstObservableNS = at
	}
	if res.UnavoidableNS == 0 && quantity >= fault.Observability.UnavoidableSNR*scan.sigma {
		res.UnavoidableNS = at
	}
	return res.UnavoidableNS != 0
}
