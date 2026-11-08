package grading

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
)

type Options = app.Options

type OutcomeOptions = app.OutcomeOptions

var ErrRunUnblinded = layer.ErrRunUnblinded

var ErrNotOneLabel = layer.ErrNotOneLabel

type ReplayMismatchError = layer.ReplayMismatchError
