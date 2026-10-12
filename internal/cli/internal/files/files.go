// Package files is the file-system edge of the cli module: the reads and
// writes that command-line paths name. Error text is the operating system's;
// callers add their own context.
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

// Write replaces the file at path with data, readable by the owner only.
func Write(path string, data []byte) error {
	// #nosec G703 -- a CLI output path is user intent.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
