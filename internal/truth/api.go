package truth

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

// Injection is one scenario's fault injection: the facts a sealed label is
// built from.
type Injection = layer.Injection

// ErrNoSpec is returned by NewSolver when no domain spec is given: the
// solver builds oracle worlds from the spec.
var ErrNoSpec = errors.New("truth: a domain spec is required")
