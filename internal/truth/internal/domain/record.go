package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Injection is one scenario's fault injection: the facts a sealed label is
// built from.
type Injection struct {
	ScenarioID    string
	Seed          uint64
	EntityID      string
	FaultID       string
	OnsetNS       int64
	StartNS       int64
	EntityIDs     []string
	PreDegraded   bool
	Perturbations []string
	Setup         []model.SetupCall
}

// Errors a facade re-exports so callers can test them with errors.Is.
var (
	ErrNoSpec   = errors.New("truth: a domain spec is required")
	ErrNoSolver = errors.New("truth: a solver is required")
	ErrNoStore  = errors.New("truth: a store is required")
)

// BuildRecord assembles one sealed label for a scenario.
func BuildRecord(spec *domain.Compiled, solver *Solver, in Injection) (*model.GroundTruthRecord, error) {
	fault, err := faultOf(spec, solver, in.FaultID)
	if err != nil {
		return nil, err
	}
	res, err := solver.solve(in.EntityID, in.FaultID, in.OnsetNS, in.StartNS, in.EntityIDs, in.Setup)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	rec := recordIdentity(spec.Spec.ID, in.ScenarioID, in.Seed, in.EntityID, in.FaultID)
	attachObservability(rec, fault, res, in.OnsetNS)
	attachScenarioContext(rec, fault, in.PreDegraded, in.Perturbations)
	return rec, nil
}

// faultOf resolves the injected fault, refusing a build with no spec or no
// solver first.
func faultOf(spec *domain.Compiled, solver *Solver, faultID string) (*model.Fault, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	if solver == nil {
		return nil, ErrNoSolver
	}
	fault := spec.Fault(faultID)
	if fault == nil {
		return nil, fmt.Errorf("truth: unknown fault %q", faultID)
	}
	return fault, nil
}

func recordIdentity(domain, scenario string, seed uint64, entity, fault string) *model.GroundTruthRecord {
	return &model.GroundTruthRecord{
		ScenarioID: scenario,
		Domain:     domain,
		Seed:       seed,
		EntityID:   entity,
		Label:      fault,
	}
}

func attachObservability(rec *model.GroundTruthRecord, fault *model.Fault, res *result, onsetNS int64) {
	rec.InjectionTimeNS = onsetNS
	rec.FirstObservableTimeNS = res.FirstObservableNS
	rec.UnavoidableTimeNS = res.UnavoidableNS
	rec.Observability = model.ObservabilityInfo{
		DetectorForm:       fault.Observability.Detector.Form,
		Channels:           res.Channels,
		EffectiveSigma:     res.EffectiveSigma,
		FirstObservableSNR: fault.Observability.FirstObservableSNR,
		UnavoidableSNR:     fault.Observability.UnavoidableSNR,
		SolutionMethod:     res.Method,
	}
}

func attachScenarioContext(rec *model.GroundTruthRecord, fault *model.Fault, preDegraded bool, perturbations []string) {
	rec.IsNegativeClass = fault.IsNegativeClass
	rec.ExpectedEpisode = !fault.IsNegativeClass
	rec.TrivialBaselineVerdict = model.TrivialNonTrivial
	rec.PreDegraded = preDegraded
	rec.Perturbations = slices.Clone(perturbations)
	rec.ExpectedEffector = fault.ExpectedEffector
	if fault.DeadlineS > 0 && rec.FirstObservableTimeNS > 0 {
		rec.DeadlineNS = rec.FirstObservableTimeNS + int64(fault.DeadlineS*1e9)
	}
	attachCounterfactual(rec, fault.Counterfactual)
}

func attachCounterfactual(rec *model.GroundTruthRecord, counterfactual *model.Counterfactual) {
	if counterfactual != nil {
		rec.Counterfactual = &model.Counterfactual{
			IfNoAction:         counterfactual.IfNoAction,
			IfActionByDeadline: counterfactual.IfActionByDeadline,
		}
	}
}
