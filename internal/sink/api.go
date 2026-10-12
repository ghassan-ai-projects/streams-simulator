package sink

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/domain"
)

type Sink = layer.Sink

var ErrNotConstructed = errors.New("sink: not built by its constructor")
