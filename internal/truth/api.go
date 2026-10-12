package truth

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/truth/internal/domain"
)

type Injection = layer.Injection

var (
	ErrNoSpec   = layer.ErrNoSpec
	ErrNoSolver = layer.ErrNoSolver
	ErrNoStore  = layer.ErrNoStore
)
