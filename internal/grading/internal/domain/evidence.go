package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// ArtifactEvidence is the evidence the artifact and its replay state on their
// own: identity, domain, effector calls, perturbations, counts and the
// reproducible and unblinded stamps. The verdict, ledger and history come from
// the files beside the artifact.
func ArtifactEvidence(art *model.RunArtifact, spec *domain.Compiled, calls []world.EffectorCall) score.Evidence {
	return score.Evidence{
		RunID: art.RunID, Domain: spec, Calls: calls, Perturbations: art.AppliedPerturbations,
		Emitted: art.Counts.Emitted, Reproducible: art.Reproducible, Unblinded: art.Unblinded,
	}
}
