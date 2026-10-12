// Package files is the file-system edge of the suite module: the directory and
// file a generated suite is written to.
package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write creates dir (and its parents, owner-only) and writes data to the file
// called name in it, readable by the owner only. It returns the file's path.
func Write(dir, name string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("%w", err)
	}
	path := filepath.Join(dir, name)
	// #nosec G703 -- a CLI output path is user intent.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return path, nil
}
