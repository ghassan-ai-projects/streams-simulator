// Package domain holds the grading rules: how the files a run published decode
// into evidence, how that evidence is assembled for the scorer, and when a
// replay is too different from the artifact to be graded.
package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/run"

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

// RequireReproduced passes a replay that reproduced the artifact and refuses
// any other with a *ReplayMismatchError.
func RequireReproduced(replayed *run.ReplayEvidence) error {
	if !replayed.Result.Matches {
		return &ReplayMismatchError{Result: replayed.Result}
	}
	return nil
}
