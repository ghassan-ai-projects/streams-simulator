package adapter

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
)

// ErrNoAdapter is returned by NewEngine when no adapter is given.
var ErrNoAdapter = errors.New("adapter: an adapter is required")

// VerifyResult is the outcome of verifying one adapter against its declared
// output schema and golden file.
type VerifyResult = layer.VerifyResult
