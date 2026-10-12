// Package files is the file-system edge of the grading module: the reads of the
// files a run published beside its artifact.
package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel"

func Read(path string) ([]byte, error) {
	return kernel.Read(path)
}
