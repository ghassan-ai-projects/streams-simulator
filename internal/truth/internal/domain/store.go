package domain

import (
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Store struct {
	mu        sync.Mutex
	labels    map[string]*model.GroundTruthRecord
	sealed    map[string]bool
	unblinded map[string]bool
	runIsOpen func(runID string) bool
}

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

func (s *Store) Seal(runID string, rec *model.GroundTruthRecord) error {
	if err := s.checkSealInput(runID, rec); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.record(runID, rec)
}

func (s *Store) checkSealInput(runID string, rec *model.GroundTruthRecord) error {
	if s == nil {
		return ErrNoStore
	}
	return validateSealInput(runID, rec)
}

func (s *Store) record(runID string, rec *model.GroundTruthRecord) error {
	if s.sealed[runID] {
		return fmt.Errorf("truth: run %q is already sealed", runID)
	}
	s.labels[runID] = cloneRecord(rec)
	s.sealed[runID] = true
	return nil
}

func (s *Store) Reveal(runID string, unblind bool) (*model.GroundTruthRecord, error) {
	if s == nil {
		return nil, ErrNoStore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reveal(runID, unblind)
}

func (s *Store) reveal(runID string, unblind bool) (*model.GroundTruthRecord, error) {
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

func (s *Store) SealStatus(runID string) (sealed, unblinded bool, err error) {
	if s == nil {
		return false, false, ErrNoStore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.labels[runID]; !ok {
		return false, false, fmt.Errorf("truth: unknown run %q", runID)
	}
	return s.sealed[runID], s.unblinded[runID], nil
}

func cloneRecord(in *model.GroundTruthRecord) *model.GroundTruthRecord {
	out := *in
	out.Observability.Channels = append([]string(nil), in.Observability.Channels...)
	out.Perturbations = append([]string(nil), in.Perturbations...)
	out.TrivialBaselineDetail = make(map[string]float64, len(in.TrivialBaselineDetail))
	// determinism-safe: copies a map into a map.
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
