package truth

// The sealed truth store. A label is sealed at run begin and revealed only
// after the run closes, or earlier with unblind:true, which stamps the run
// permanently and excludes it from every scorecard.

import (
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Store holds the sealed truth for runs under the director role. It is
// deliberately a separate object from any operator-facing view.
type Store struct {
	mu        sync.Mutex
	labels    map[string]*model.GroundTruthRecord // by run id
	sealed    map[string]bool
	unblinded map[string]bool
	runIsOpen func(runID string) bool
}

// NewStore builds an empty truth store. runIsOpen reports whether a run is
// still open; Reveal refuses an open run unless it is unblinded. The check is
// required: a nil check counts every run as open, so a store built without
// one refuses to reveal rather than leaking the label.
func NewStore(runIsOpen func(runID string) bool) *Store {
	if runIsOpen == nil {
		runIsOpen = func(string) bool { return true }
	}
	return &Store{
		runIsOpen: runIsOpen,
		labels:    map[string]*model.GroundTruthRecord{},
		sealed:    map[string]bool{},
		unblinded: map[string]bool{},
	}
}

// Seal records the label for a run and seals it.
func (s *Store) Seal(runID string, rec *model.GroundTruthRecord) error {
	if err := validateSealInput(runID, rec); err != nil {
		return err
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
	if s.runIsOpen(runID) && !unblind && !s.unblinded[runID] {
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

func validateSealInput(runID string, rec *model.GroundTruthRecord) error {
	if runID == "" {
		return fmt.Errorf("truth: run id is required")
	}
	if rec == nil {
		return fmt.Errorf("truth: label is required")
	}
	return nil
}
