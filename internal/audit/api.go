package audit

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/audit/internal/domain"
)

type Verdict = layer.Verdict

type Perturbation = layer.Perturbation

var ErrNoSpec = errors.New("audit: a domain spec is required")

var ErrNoPanel = errors.New("audit: a panel built by NewPanel is required")
