package audit

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Audit evaluates one injection against the panel and reports whether a
// hindsight-fitted trivial detector separates it from its control.
func (p *Panel) Audit(entityID, faultID string, onsetNS, startNS int64, entityIDs []string, durationNS int64, setup []model.SetupCall, perturbations []Perturbation) (*Verdict, error) {
	if p == nil || p.panel == nil {
		return nil, ErrNoPanel
	}
	return p.panel.Audit(entityID, faultID, onsetNS, startNS, entityIDs, durationNS, setup, perturbations)
}
