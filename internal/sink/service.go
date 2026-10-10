package sink

import (
	"context"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/httppush"
)

// Inproc buffers everything in memory. The zero value is ready to use.
type Inproc struct {
	impl layer.Inproc
}

// File writes lines to a path, buffered; Close flushes.
type File struct {
	file *files.File
}

// NewFile opens the sink at path (created if missing, with its directories).
func NewFile(path string) (*File, error) {
	inner, err := files.New(path)
	if err != nil {
		return nil, err
	}
	return &File{file: inner}, nil
}

// HTTPPush posts each line to an endpoint. In stepped mode delivery is
// immediate and in order, so the sink is byte-deterministic.
type HTTPPush struct {
	push *httppush.HTTPPush
}

// NewHTTPPush builds the sink. ctx cancels the run.
func NewHTTPPush(ctx context.Context, url string) *HTTPPush {
	return &HTTPPush{push: httppush.New(ctx, url)}
}
