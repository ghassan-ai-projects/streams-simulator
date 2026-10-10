package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func defaultEntities(spec *domain.Compiled, prof *model.Profile) []string {
	n := prof.EntityCount
	if n <= 0 {
		n = spec.Spec.Entities.Count.Default
	}
	if n <= 0 {
		n = 1
	}
	var out []string
	for i := 1; i <= n; i++ {
		out = append(out, world.RenderID(spec.Spec.Entities.IDTemplate, i))
	}
	return out
}
