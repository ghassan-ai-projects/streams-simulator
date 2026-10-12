package perturb

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

type Delivered = layer.Delivered

var ErrNoSpec = errors.New("perturb: a domain spec is required")
