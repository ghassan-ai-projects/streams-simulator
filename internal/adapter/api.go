package adapter

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
)

var ErrNoAdapter = errors.New("adapter: an adapter is required")

var ErrUnknownAdapter = layer.ErrUnknownAdapter

type VerifyResult = layer.VerifyResult
