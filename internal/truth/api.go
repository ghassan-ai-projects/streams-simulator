package truth

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

// Injection is one scenario's fault injection: the facts a sealed label is
// built from.
type Injection = layer.Injection

// Errors returned for a missing dependency: NewSolver and BuildRecord need a
// domain spec, BuildRecord a solver, and a zero Store has no backing store.
var (
	ErrNoSpec   = layer.ErrNoSpec
	ErrNoSolver = layer.ErrNoSolver
	ErrNoStore  = layer.ErrNoStore
)
