// Package files is the file-system edge of the plantwire module: it reads kind files and lists kind directories.
package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func Scan(dirs []string) ([]domain.KindSource, error) {
	var sources []domain.KindSource
	for _, dir := range dirs {
		paths, err := kindPaths(dir)
		if err != nil {
			return nil, err
		}
		sources = append(sources, readEach(paths)...)
	}
	return sources, nil
}

func ScanPresent(dirs []string) ([]domain.KindSource, error) {
	var sources []domain.KindSource
	for _, dir := range dirs {
		found, err := Scan([]string{dir})
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		sources = append(sources, found...)
	}
	return sources, nil
}

func readEach(paths []string) []domain.KindSource {
	sources := make([]domain.KindSource, len(paths))
	for i, path := range paths {
		raw, err := kernel.ReadRegular(path, MaxKindFileBytes)
		if err != nil {
			err = fmt.Errorf("plantwire: read %s: %w", path, err)
		}
		sources[i] = domain.KindSource{Path: path, Raw: raw, Err: err}
	}
	return sources
}

func directory(dir string) ([]domain.KindSource, error) {
	paths, err := kindPaths(dir)
	if err != nil {
		return nil, err
	}
	return Paths(paths)
}

func kindPaths(dir string) ([]string, error) {
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
	return paths, nil
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

func OpenCapturable(path string) (io.ReadCloser, error) {
	file, err := kernel.OpenRegular(path)
	if err != nil {
		return nil, fmt.Errorf("plantwire: read %s: %w", path, err)
	}
	return file, nil
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
