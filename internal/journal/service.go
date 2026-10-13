package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/files"
)

type Journal struct {
	impl *layer.Journal
}

func NewMemory(retain bool) *Journal {
	return &Journal{impl: layer.NewMemory(retain)}
}

func NewInDir(dir string, retain bool) (*Journal, error) {
	inner, err := files.Open(dir, retain)
	if err != nil {
		return nil, err
	}
	return &Journal{impl: inner}, nil
}
