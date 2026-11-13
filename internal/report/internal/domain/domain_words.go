package domain

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func embeddedDomain(artifact *model.RunArtifact) (*model.DomainSpec, bool) {
	if artifact == nil || len(artifact.DomainSpec) == 0 {
		return nil, false
	}
	var spec model.DomainSpec
	if json.Unmarshal(artifact.DomainSpec, &spec) != nil {
		return nil, false
	}
	return &spec, true
}

func outcomeWords(artifact *model.RunArtifact) (unit string, labels map[string]string) {
	spec, ok := embeddedDomain(artifact)
	if !ok {
		return "", nil
	}
	if weights := spec.GroundTruth.OutcomeWeights; weights != nil {
		unit = weights.Unit
	}
	for _, policy := range spec.GroundTruth.BaselinePolicies {
		if policy.Label != "" {
			labels = withLabel(labels, string(model.PolicyOutcome(policy.Name)), policy.Label)
		}
	}
	return unit, labels
}

func withLabel(labels map[string]string, mode, label string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[mode] = label
	return labels
}
