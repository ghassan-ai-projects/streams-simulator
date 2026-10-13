package audit

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (p *Panel) Audit(entityID, faultID string, onsetNS, startNS int64, entityIDs []string, durationNS int64, setup []model.SetupCall, perturbations []Perturbation) (*Verdict, error) {
	if p == nil || p.panel == nil {
		return nil, ErrNoPanel
	}
	return p.panel.Audit(entityID, faultID, onsetNS, startNS, entityIDs, durationNS, setup, perturbations)
}
