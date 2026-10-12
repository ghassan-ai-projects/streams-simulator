package sink

import (
	"context"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink/internal/httppush"
)

type Inproc struct {
	impl layer.Inproc
}

type File struct {
	file *files.File
}

func NewFile(path string) (*File, error) {
	inner, err := files.New(path)
	if err != nil {
		return nil, err
	}
	return &File{file: inner}, nil
}

type HTTPPush struct {
	push *httppush.HTTPPush
}

func NewHTTPPush(ctx context.Context, url string) *HTTPPush {
	return &HTTPPush{push: httppush.New(ctx, url)}
}
