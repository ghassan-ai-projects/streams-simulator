package grading

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
)

// Options name the run to grade: its artifact, the sealed label (empty selects
// label.json beside the artifact), and the directories the artifact's domain
// and adapter are looked up in to replay it.
type Options = app.Options

// ErrRunUnblinded is the refusal to grade a run stamped unblinded (an explore
// run, or one whose truth was revealed while it was open).
var ErrRunUnblinded = layer.ErrRunUnblinded

// ReplayMismatchError is the refusal to grade a run whose replay does not
// reproduce its recorded trace; Result is the replay report.
type ReplayMismatchError = layer.ReplayMismatchError
