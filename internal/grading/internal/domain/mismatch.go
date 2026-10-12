// Package domain holds the grading rules: how the files a run published decode
// into evidence, how that evidence is assembled for the scorer, and when a
// replay is too different from the artifact to be graded.
package domain

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// ReplayMismatchError refuses to grade a run whose replay does not reproduce
// the recorded trace, because its effector calls would not be the ones the run
// made. Result is the replay report.
type ReplayMismatchError struct {
	Result *run.ReplayResult
}

// Error states the refusal and the first divergence.
func (e *ReplayMismatchError) Error() string {
	return "score: replay does not reproduce the artifact, so its effector calls cannot be graded: " + e.Result.Detail
}

// ErrRunUnblinded refuses to grade a run stamped unblinded (an explore run, or
// one whose truth was revealed while it was open): an unblinded run is excluded
// from every scorecard.
var ErrRunUnblinded = errors.New("score: the run is stamped unblinded and is never scored")

// RequireBlind passes a run that was not unblinded and refuses any other with
// ErrRunUnblinded. An explore run is refused by its mode as well as by its
// stamp, so clearing the flag in the artifact does not make it scorable.
func RequireBlind(art *model.RunArtifact) error {
	if art.Unblinded || art.Mode == model.ModeExplore {
		return ErrRunUnblinded
	}
	return nil
}

// RequireReproduced passes a replay that reproduced the artifact and refuses
// any other with a *ReplayMismatchError.
func RequireReproduced(replayed *run.ReplayEvidence) error {
	if !replayed.Result.Matches {
		return &ReplayMismatchError{Result: replayed.Result}
	}
	return nil
}
