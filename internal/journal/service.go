package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/files"
)

// Journal is the pair of streams of one run.
type Journal struct {
	impl *layer.Journal
}

// NewMemory is a journal with no files: it keeps the digests of both streams
// and, with retain, their bytes.
func NewMemory(retain bool) *Journal {
	return &Journal{impl: layer.NewMemory(retain)}
}

// NewInDir creates dir/journal.jsonl and dir/sealed/journal.jsonl and returns
// the journal writing to them; retain also keeps the bytes in memory.
func NewInDir(dir string, retain bool) (*Journal, error) {
	inner, err := files.Open(dir, retain)
	if err != nil {
		return nil, err
	}
	return &Journal{impl: inner}, nil
}
