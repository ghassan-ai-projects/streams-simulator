package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func OutcomeModes(spec *model.DomainSpec) []model.OutcomeMode {
	modes := []model.OutcomeMode{model.OutcomeAsRun, model.OutcomeDoNothing}
	for _, policy := range spec.GroundTruth.BaselinePolicies {
		modes = append(modes, model.PolicyOutcome(policy.Name))
	}
	return modes
}

func RequireMatch(result *run.ReplayResult) error {
	if !run.Reproduced(result) {
		return &ReplayMismatchError{Result: result}
	}
	return nil
}
