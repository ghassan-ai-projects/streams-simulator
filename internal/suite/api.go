package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
)

type Perturbation = audit.Perturbation

type PublishOptions = app.PublishOptions

type Published = app.Published

type UnknownProfileError = layer.UnknownProfileError
