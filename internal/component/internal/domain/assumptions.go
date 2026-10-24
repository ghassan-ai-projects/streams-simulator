package domain

import (
	"encoding/json"
	"fmt"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func DescribeAssumptions(lib *Library, input AssumptionsInput) (*Assumptions, error) {
	compiled, err := specs.Parse(input.Domain, input.Source)
	if err != nil {
		return nil, fmt.Errorf("component: %w", err)
	}
	overlay, err := overlayOf(lib, compiled, input.Twin)
	if err != nil {
		return nil, err
	}
	return validatedAssumptions(assemble(compiled, overlay))
}

func validatedAssumptions(assumptions *Assumptions) (*Assumptions, error) {
	raw, err := json.Marshal(assumptions)
	if err != nil {
		return nil, fmt.Errorf("component: encode assumptions: %w", err)
	}
	if err := model.ValidateDomainAssumptions(raw); err != nil {
		return nil, fmt.Errorf("component: %w", err)
	}
	return assumptions, nil
}

func assemble(compiled *specs.Compiled, overlay *twinOverlay) *Assumptions {
	spec := compiled.Spec
	assumptions := &Assumptions{SchemaVersion: assumptionsVersion, Domain: DomainIdentity{ID: spec.ID, Version: spec.Version, Digest: compiled.Digest},
		Origin: originOf(spec, overlay), Kinds: make([]KindAssumptions, 0, len(spec.Kinds))}
	for i := range spec.Kinds {
		assumptions.Kinds = append(assumptions.Kinds, kindAssumptions(spec, &spec.Kinds[i], overlay))
	}
	assumptions.NotModeled = notModeled(assumptions)
	return assumptions
}

func originOf(spec *model.DomainSpec, overlay *twinOverlay) AssumptionsOrigin {
	switch {
	case overlay.inSync():
		return AssumptionsOrigin{Kind: OriginTwin, Twin: overlay.origin}
	case overlay != nil:
		return AssumptionsOrigin{Kind: OriginDomain, Twin: overlay.origin}
	case hasProvenance(spec):
		return AssumptionsOrigin{Kind: OriginComposition}
	}
	return AssumptionsOrigin{Kind: OriginDomain}
}

func hasProvenance(spec *model.DomainSpec) bool {
	for i := range spec.Instances {
		if spec.Instances[i].Provenance != nil {
			return true
		}
	}
	return false
}

func kindAssumptions(spec *model.DomainSpec, kind *model.Kind, overlay *twinOverlay) KindAssumptions {
	return KindAssumptions{Kind: kind.Name, Implements: kind.Implements, Instances: instanceAssumptions(spec, kind.Name),
		Readings: readingAssumptions(spec, kind, overlay), Dynamics: dynamicsAssumptions(kind), Commands: commandAssumptions(kind, overlay), Faults: faultAssumptions(kind)}
}

func instanceAssumptions(spec *model.DomainSpec, kind string) []InstanceAssumption {
	sets := []InstanceAssumption{}
	for _, set := range spec.Instances {
		if set.Kind == kind {
			sets = append(sets, InstanceAssumption{Names: set.Names, IDTemplate: set.IDTemplate, CountDefault: set.Count.Default, Area: set.Area})
		}
	}
	return sets
}
