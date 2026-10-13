package component

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
)

type Summary = layer.Summary

type Detail = layer.Detail

var ErrNoLibrary = errors.New("component: a library built by NewLibrary is required")
