// Package files is the file-system edge of the grading module: the reads of the
// files a run published beside its artifact.
package files

import (
	"fmt"
	"os"
)

// Read returns the contents of the file at path.
func Read(path string) ([]byte, error) {
	// #nosec G304 G703 -- a CLI flag naming a file is user intent.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
}
