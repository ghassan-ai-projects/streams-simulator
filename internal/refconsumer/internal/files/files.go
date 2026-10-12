// Package files is the file-system edge of the refconsumer module: it reads
// the trace the reference consumer detects in and writes the verdict it
// reports.
package files

import (
	"fmt"
	"os"
)

// ReadTrace returns the contents of the native-format trace at path.
func ReadTrace(path string) ([]byte, error) {
	// #nosec G304 G703 -- a CLI flag naming a file is user intent.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
}

// WriteVerdict replaces the file at path with data, readable by the owner only.
func WriteVerdict(path string, data []byte) error {
	// #nosec G703 -- a CLI output path is user intent.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
