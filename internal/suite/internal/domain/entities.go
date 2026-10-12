package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func defaultEntities(spec *domain.Compiled, prof *model.Profile) []string {
	return world.PopulationIDs(spec, prof.EntityCount)
}

func (g *suiteGeneration) faultEntities(fault string) []string {
	if g.cfg.Domain.FaultKind(fault) == nil {
		return g.entities
	}
	var out []string
	for _, id := range g.entities {
		if kindOfEntity(g.cfg.Domain, id).FaultDecl(fault) != nil {
			out = append(out, id)
		}
	}
	return out
}

func kindOfEntity(spec *domain.Compiled, id string) *model.Kind {
	return spec.InstanceKind(max(spec.InstanceSetFor(id), 0))
}

func declaredFaults(spec *domain.Compiled) []*model.Fault {
	var faults []*model.Fault
	for k := range spec.Spec.Kinds {
		for i := range spec.Spec.Kinds[k].Faults {
			faults = append(faults, &spec.Spec.Kinds[k].Faults[i])
		}
	}
	return faults
}
