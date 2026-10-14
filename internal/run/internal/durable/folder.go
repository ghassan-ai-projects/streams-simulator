package durable

import (
	"errors"
	"io/fs"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func FolderUsed(dir string) bool {
	_, err := kernel.Names(dir)
	return !errors.Is(err, fs.ErrNotExist)
}
