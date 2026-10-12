package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func lower(document *model.DomainDocument, src string) (*model.DomainSpec, error) {
	kind, instances := &document.Kinds[0], &document.Instances[0]
	if instances.Kind != kind.Name {
		return nil, invalidDomain(src, "/instances/0/kind", fmt.Sprintf("instances name kind %q, but the declared kind is %q", instances.Kind, kind.Name))
	}
	return &model.DomainSpec{
		ID: document.ID, Version: document.Version, Title: document.Title, Description: document.Description,
		Axes: document.Axes, Stresses: document.Stresses, Entities: lowerEntities(instances),
		State: kind.State, Dynamics: kind.Dynamics, Channels: kind.Channels, Faults: kind.Faults, Effectors: kind.Effectors,
		Profiles: document.Profiles, GroundTruth: document.GroundTruth,
	}, nil
}

func lowerEntities(instances *model.InstanceSet) model.Entities {
	return model.Entities{IDTemplate: instances.IDTemplate, EntityType: instances.Kind, Count: instances.Count, Churn: instances.Churn}
}
