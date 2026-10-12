package adapter

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
)

// ErrNoAdapter is returned by NewEngine when no adapter is given.
var ErrNoAdapter = errors.New("adapter: an adapter is required")

// ErrUnknownAdapter is returned by Lookup for an id the set does not hold.
var ErrUnknownAdapter = layer.ErrUnknownAdapter

// VerifyResult is the outcome of verifying one adapter against its declared
// output schema and golden file.
type VerifyResult = layer.VerifyResult
