package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"

func FolderUsed(dir string) bool {
	return durable.FolderUsed(dir)
}
