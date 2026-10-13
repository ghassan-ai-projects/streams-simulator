package files

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const builtInLabel = "builtin:"

type tree interface {
	names(dir string) ([]string, error)
	read(path string) ([]byte, error)
	readIfPresent(path string) ([]byte, bool, error)
	join(parts ...string) string
	label(path string) string
}

type diskTree struct{}

func (diskTree) names(dir string) ([]string, error) { return kernel.Names(dir) }

func (diskTree) read(file string) ([]byte, error) { return kernel.Read(file) }

func (diskTree) readIfPresent(file string) ([]byte, bool, error) { return kernel.ReadIfPresent(file) }

func (diskTree) join(parts ...string) string { return filepath.Join(parts...) }

func (diskTree) label(file string) string { return file }

type embeddedTree struct {
	files fs.FS
	root  string
}

func (t embeddedTree) names(dir string) ([]string, error) {
	entries, err := fs.ReadDir(t.files, path.Join(t.root, dir))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func (t embeddedTree) read(file string) ([]byte, error) {
	raw, err := fs.ReadFile(t.files, path.Join(t.root, file))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
}

func (t embeddedTree) readIfPresent(file string) ([]byte, bool, error) {
	raw, err := t.read(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return raw, err == nil, err
}

func (embeddedTree) join(parts ...string) string { return path.Join(parts...) }

func (embeddedTree) label(file string) string { return builtInLabel + file }
