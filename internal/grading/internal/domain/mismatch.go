// Package domain holds the grading rules: how the files a run published decode
// into evidence, how that evidence is assembled for the scorer, and when a
// replay is too different from the artifact to be graded.
package domain

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type ReplayMismatchError struct {
	Result *run.ReplayResult
}

func (e *ReplayMismatchError) Error() string {
	return "score: replay does not reproduce the artifact, so its effector calls cannot be graded: " + e.Result.Detail
}

var ErrRunUnblinded = errors.New("score: the run is stamped unblinded and is never scored")

func RequireBlind(art *model.RunArtifact) error {
	if art.Unblinded || art.Mode == model.ModeExplore {
		return ErrRunUnblinded
	}
	return nil
}

func RequireReproduced(replayed *run.ReplayEvidence) error {
	if !run.Reproduced(replayed.Result) {
		return &ReplayMismatchError{Result: replayed.Result}
	}
	return nil
}
