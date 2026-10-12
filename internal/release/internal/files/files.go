// Package files is the file-system edge of the release module: the directories
// whose files a release digests, the signing key file and the manifest file.
package files

import (
	"fmt"
	"os"
)

// Names lists the names of the non-directory entries of dir, sorted by name.
func Names(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

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
