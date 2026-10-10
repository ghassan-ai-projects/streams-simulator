package app

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// artifact assembles the run artifact from the run state.
func (r *Run) artifact() *model.RunArtifact {
	commands := r.sortedCommandLog()
	counts := r.artifactCounts()
	artifact := r.artifactIdentity()
	r.attachArtifactInputs(artifact)
	r.attachArtifactState(artifact, commands, counts)
	return artifact
}
