package truth

// The ground truth record and the truth store. Sealed at run begin;
// revealed only after the run closes, or earlier with unblind:true, which
// stamps the run permanently and excludes it from every scorecard.

import (
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// BuildRecord assembles one sealed label for a scenario.
func BuildRecord(
	spec *domain.Compiled,
	solver *Solver,
	scenarioID string,
	seed uint64,
	entityID, faultID string,
	onsetNS, startNS int64,
	entityIDs []string,
	preDegraded bool,
	perturbations []string,
	setup []SetupCall,
) (*model.GroundTruthRecord, error) {
	fault := spec.Fault(faultID)
	if fault == nil {
		return nil, fmt.Errorf("truth: unknown fault %q", faultID)
	}
	res, err := solver.Solve(entityID, faultID, onsetNS, startNS, entityIDs, setup)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	rec := &model.GroundTruthRecord{
		ScenarioID:            scenarioID,
		Domain:                spec.Spec.ID,
		Seed:                  seed,
		EntityID:              entityID,
		Label:                 faultID,
		IsNegativeClass:       fault.IsNegativeClass,
		ExpectedEpisode:       !fault.IsNegativeClass,
		InjectionTimeNS:       onsetNS,
		FirstObservableTimeNS: res.FirstObservableNS,
		UnavoidableTimeNS:     res.UnavoidableNS,
		Observability: model.ObservabilityInfo{
			DetectorForm:       fault.Observability.Detector.Form,
			Channels:           res.Channels,
			EffectiveSigma:     res.EffectiveSigma,
			FirstObservableSNR: fault.Observability.FirstObservableSNR,
			UnavoidableSNR:     fault.Observability.UnavoidableSNR,
			SolutionMethod:     res.Method,
		},
		TrivialBaselineVerdict: model.TrivialNonTrivial,
		PreDegraded:            preDegraded,
		Perturbations:          perturbations,
	}
	if fault.ExpectedEffector != "" {
		rec.ExpectedEffector = fault.ExpectedEffector
	}
	if fault.DeadlineS > 0 {
		if res.FirstObservableNS > 0 {
			rec.DeadlineNS = res.FirstObservableNS + int64(fault.DeadlineS*1e9)
		}
	}
	if fault.Counterfactual != nil {
		rec.Counterfactual = &model.Counterfactual{
			IfNoAction:         fault.Counterfactual.IfNoAction,
			IfActionByDeadline: fault.Counterfactual.IfActionByDeadline,
		}
	}
	return rec, nil
}

// Store holds the sealed truth for runs under the director role. It is
// deliberately a separate object from any operator-facing view.
type Store struct {
	mu        sync.Mutex
	labels    map[string]*model.GroundTruthRecord // by run id
	sealed    map[string]bool
	unblinded map[string]bool
	// Verification hooks: reveal refusal on an open run.
	OpenChecker func(runID string) bool
}

// NewStore builds an empty truth store.
func NewStore() *Store {
	return &Store{
		labels:    map[string]*model.GroundTruthRecord{},
		sealed:    map[string]bool{},
		unblinded: map[string]bool{},
	}
}

// Seal records the label for a run and seals it.
func (s *Store) Seal(runID string, rec *model.GroundTruthRecord) error {
	if runID == "" {
		return fmt.Errorf("truth: run id is required")
	}
	if rec == nil {
		return fmt.Errorf("truth: label is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sealed[runID] {
		return fmt.Errorf("truth: run %q is already sealed", runID)
	}
	s.labels[runID] = cloneRecord(rec)
	s.sealed[runID] = true
	return nil
}

// Reveal returns the sealed label. unblind permits revealing on an open
// run, permanently stamping it.
func (s *Store) Reveal(runID string, unblind bool) (*model.GroundTruthRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.labels[runID]
	if !ok {
		return nil, fmt.Errorf("truth: no sealed label for run %q", runID)
	}
	if s.OpenChecker != nil && s.OpenChecker(runID) && !unblind && !s.unblinded[runID] {
		return nil, fmt.Errorf("truth: reveal refused on an open run (call with unblind:true to stamp and reveal)")
	}
	if unblind {
		s.unblinded[runID] = true
	}
	return cloneRecord(rec), nil
}

// SealStatus reports the sealing state.
func (s *Store) SealStatus(runID string) (sealed, unblinded bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.labels[runID]; !ok {
		return false, false, fmt.Errorf("truth: unknown run %q", runID)
	}
	return s.sealed[runID], s.unblinded[runID], nil
}

// cloneRecord returns a defensive copy so callers cannot mutate the sealed
// oracle through shared slices, maps, or pointers.
func cloneRecord(in *model.GroundTruthRecord) *model.GroundTruthRecord {
	out := *in
	out.Observability.Channels = append([]string(nil), in.Observability.Channels...)
	out.Perturbations = append([]string(nil), in.Perturbations...)
	out.TrivialBaselineDetail = make(map[string]float64, len(in.TrivialBaselineDetail))
	for k, v := range in.TrivialBaselineDetail {
		out.TrivialBaselineDetail[k] = v
	}
	if in.Counterfactual != nil {
		cf := *in.Counterfactual
		out.Counterfactual = &cf
	}
	return &out
}
