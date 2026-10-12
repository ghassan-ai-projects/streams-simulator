// Package domain holds the trivial-baseline audit rules: a panel of one-line
// detectors fitted with hindsight on the scenario itself, deliberately unfair
// to the scenario. If a detector tuned on the answer still cannot separate
// the fault from its control at >= 0.9 balanced accuracy, the scenario is
// non_trivial and may enter the graded suite. It performs no I/O.
package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const BalancedAccuracyCutoff = 0.9

var DetectorNames = []string{
	"fixed_threshold", "zscore", "first_difference", "moving_median_residual", "channel_silence",
}

type Verdict struct {
	Trivial   bool               `json:"trivial"`
	Scores    map[string]float64 `json:"scores"`
	Best      string             `json:"best"`
	BestScore float64            `json:"best_score"`
	Channels  []string           `json:"channels"`
	Samples   int                `json:"samples"`
}

type Panel struct {
	spec     *domain.Compiled
	seed     uint64
	sampleNS int64
}

func NewPanel(spec *domain.Compiled, seed uint64, sampleNS int64) *Panel {
	if sampleNS <= 0 {
		sampleNS = 60 * 1e9
	}
	return &Panel{spec: spec, seed: seed, sampleNS: sampleNS}
}

type Perturbation struct {
	Name    string         `json:"name"`
	Params  map[string]any `json:"params,omitempty"`
	FromNS  int64          `json:"from_ns,omitempty"`
	UntilNS int64          `json:"until_ns,omitempty"`
}

func (p *Panel) Audit(entityID, faultID string, onsetNS, startNS int64, entityIDs []string, durationNS int64, setup []model.SetupCall, perturbations []Perturbation) (*Verdict, error) {
	clean, cleanEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{p.controlFault(faultID): onsetNS}, setup, perturbations, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	faulted, faultEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup, perturbations, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	evidence := p.sampleEvidence(entityID, faultID, clean, faulted, onsetNS, startNS, durationNS)
	return p.gradeEvidence(evidence, cleanEmissions, faultEmissions, startNS, durationNS), nil
}
