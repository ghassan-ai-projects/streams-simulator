package sink

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/domain"
)

// Sink receives rendered lines and can produce the full byte stream at
// Close, which is what the trace digest is computed from.
type Sink = layer.Sink

// ErrNotConstructed is returned by a File or HTTPPush that was not built by
// its constructor.
var ErrNotConstructed = errors.New("sink: not built by its constructor")
