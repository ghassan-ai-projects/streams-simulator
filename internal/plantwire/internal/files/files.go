// Package files is the file-system edge of the plantwire module: it reads kind files and lists kind directories.
package files

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const (
	MaxKindFileBytes = 1 << 20
	StreamName       = "<stdin>"
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

func Stream(reader io.Reader) (domain.KindSource, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, MaxKindFileBytes+1))
	if err != nil {
		return domain.KindSource{}, fmt.Errorf("plantwire: read %s: %w", StreamName, err)
	}
	if len(raw) > MaxKindFileBytes {
		return domain.KindSource{}, fmt.Errorf("plantwire: %s holds more than %d bytes: %w", StreamName, MaxKindFileBytes, kernel.ErrTooLarge)
	}
	return domain.KindSource{Path: StreamName, Raw: raw}, nil
}

func Paths(paths []string) ([]domain.KindSource, error) {
	sources := make([]domain.KindSource, 0, len(paths))
	for _, path := range paths {
		raw, err := kernel.ReadRegular(path, MaxKindFileBytes)
		if err != nil {
			return nil, fmt.Errorf("plantwire: read %s: %w", path, err)
		}
		sources = append(sources, domain.KindSource{Path: path, Raw: raw})
	}
	return sources, nil
}
