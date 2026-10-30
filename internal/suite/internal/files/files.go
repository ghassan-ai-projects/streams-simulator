// Package files is the file-system edge of the suite module: the directory and
// file a generated suite is written to.
package files

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func Write(dir, name string, data []byte) (string, error) {
	if err := kernel.PrivateDir(dir); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	return path, kernel.WriteAtomic(path, data)
}

const maxSpecBytes = 1 << 20

func Read(path string) ([]byte, error) {
	return kernel.ReadRegular(path, maxSpecBytes)
}

func Names(dir string) ([]string, error) {
	names, err := kernel.Names(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return names, err
}
