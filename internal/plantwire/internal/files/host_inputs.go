package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel"

func ReadHostInput(path string) ([]byte, error) {
	return kernel.Read(path)
}
