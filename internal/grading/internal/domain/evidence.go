package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func ArtifactEvidence(art *model.RunArtifact, spec *domain.Compiled, calls []world.EffectorCall) score.Evidence {
	return score.Evidence{
		RunID: art.RunID, Domain: spec, Calls: calls, Perturbations: art.AppliedPerturbations,
		Emitted: art.Counts.Emitted, Reproducible: art.Reproducible, Unblinded: art.Unblinded,
	}
}
