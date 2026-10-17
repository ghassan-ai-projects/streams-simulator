// Package files is the file-system edge of the plantwire module: it reads kind files and lists kind directories.
package files

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func Directories(dirs []string) ([]domain.KindSource, error) {
	var sources []domain.KindSource
	for _, dir := range dirs {
		found, err := directory(dir)
		if err != nil {
			return nil, err
		}
		sources = append(sources, found...)
	}
	return sources, nil
}

func directory(dir string) ([]domain.KindSource, error) {
	names, err := kernel.Names(dir)
	if err != nil {
		return nil, fmt.Errorf("plantwire: list %s: %w", dir, err)
	}
	slices.Sort(names)
	var paths []string
	for _, name := range names {
		if strings.HasSuffix(name, domain.KindFileSuffix) {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	return Paths(paths)
}

func Paths(paths []string) ([]domain.KindSource, error) {
	sources := make([]domain.KindSource, 0, len(paths))
	for _, path := range paths {
		raw, err := kernel.Read(path)
		if err != nil {
			return nil, fmt.Errorf("plantwire: read %s: %w", path, err)
		}
		sources = append(sources, domain.KindSource{Path: path, Raw: raw})
	}
	return sources, nil
}
