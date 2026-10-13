// Package files is the file-system edge of the refconsumer module: it reads
// the trace the reference consumer detects in and writes the verdict it
// reports.
package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel"

func ReadTrace(path string) ([]byte, error) {
	return kernel.Read(path)
}

func WriteVerdict(path string, data []byte) error {
	return kernel.WriteAtomic(path, data)
}
