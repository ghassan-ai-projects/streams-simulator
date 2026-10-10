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

// BalancedAccuracyCutoff is the hindsight-fitted bar for triviality.
const BalancedAccuracyCutoff = 0.9

// DetectorNames is the panel in a stable order.
var DetectorNames = []string{
	"fixed_threshold", "zscore", "first_difference", "moving_median_residual", "channel_silence",
}

// Verdict is the audit outcome for one scenario.
type Verdict struct {
	Trivial   bool               `json:"trivial"`
	Scores    map[string]float64 `json:"scores"`
	Best      string             `json:"best"`
	BestScore float64            `json:"best_score"`
	Channels  []string           `json:"channels"`
	Samples   int                `json:"samples"`
}

// Panel audits one injection: it runs the world twice (clean control and
// faulted), samples the delivered channel readings on a grid, and fits each
// trivial detector with hindsight on the labeled series.
type Panel struct {
	spec     *domain.Compiled
	seed     uint64
	sampleNS int64
}

// NewPanel builds the audit panel.
func NewPanel(spec *domain.Compiled, seed uint64, sampleNS int64) *Panel {
	if sampleNS <= 0 {
		sampleNS = 60 * 1e9
	}
	return &Panel{spec: spec, seed: seed, sampleNS: sampleNS}
}

// Perturbation is one delivery perturbation the audit applies to the
// delivered stream, exactly as the scenario declares it.
type Perturbation struct {
	Name    string         `json:"name"`
	Params  map[string]any `json:"params,omitempty"`
	FromNS  int64          `json:"from_ns,omitempty"`
	UntilNS int64          `json:"until_ns,omitempty"`
}

// Audit runs the panel on one injection. The control is the domain's
// declared negative-class scenario (transient_none-style) when one exists,
// else a clean world: the design's bar is separating the fault from its
// negative controls, not from an empty trace. setup applies pre-fault
// effector calls to both worlds (scenario context such as an aerator
// running at night). The detector inputs are the immutable delivered
// stream — world, perturbation layer and adapter projection — in monotonic
// order, never random access on a mutable world.
func (p *Panel) Audit(entityID, faultID string, onsetNS, startNS int64, entityIDs []string, durationNS int64, setup []model.SetupCall, perturbations []Perturbation) (*Verdict, error) {
	clean, cleanEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{p.controlFault(): onsetNS}, setup, perturbations, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	faulted, faultEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup, perturbations, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	evidence := p.sampleEvidence(clean, faulted, onsetNS, startNS, durationNS)
	return p.gradeEvidence(evidence, cleanEmissions, faultEmissions, startNS, durationNS), nil
}
