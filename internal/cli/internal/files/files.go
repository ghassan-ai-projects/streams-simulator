// Package files is the file-system edge of the cli module: the reads and
// writes that command-line paths name. Error text is the operating system's;
// callers add their own context.
package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel"

func Read(path string) ([]byte, error) {
	return kernel.Read(path)
}

func Write(path string, data []byte) error {
	return kernel.WriteAtomic(path, data)
}
