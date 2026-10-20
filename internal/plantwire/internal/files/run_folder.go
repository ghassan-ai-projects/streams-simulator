package files

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const sealedFolder = "sealed"

func RunFolderFiles(dir string) ([]string, error) {
	top, err := presentNames(dir)
	if err != nil {
		return nil, err
	}
	sealed, err := presentNames(filepath.Join(dir, sealedFolder))
	if err != nil {
		return nil, err
	}
	for _, name := range sealed {
		top = append(top, path.Join(sealedFolder, name))
	}
	return top, nil
}

func presentNames(dir string) ([]string, error) {
	names, err := kernel.Names(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("plantwire: list the run folder %s: %w", dir, err)
	}
	return names, nil
}
