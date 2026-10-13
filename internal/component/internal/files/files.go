// Package files is the file-system edge of the component module: it reads the
// directories of a component library and hands the bytes to the pure layer.
package files

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const vocabularyFile = "vocabulary.json"

func Load(dirs []string) (*domain.Library, error) {
	directories := make([]domain.Directory, 0, len(dirs))
	for _, dir := range dirs {
		directory, err := readDirectory(dir)
		if err != nil {
			return nil, err
		}
		directories = append(directories, directory)
	}
	return domain.LoadLibrary(directories)
}

func readDirectory(dir string) (domain.Directory, error) {
	if _, err := kernel.Names(dir); err != nil {
		return domain.Directory{}, fmt.Errorf("component: list %s: %w", dir, err)
	}
	equipment, err := readGroup(dir, "equipment", ".component.json")
	if err != nil {
		return domain.Directory{}, err
	}
	instruments, err := readGroup(dir, "instruments", ".instrument.json")
	if err != nil {
		return domain.Directory{}, err
	}
	vocabulary, err := readVocabulary(dir)
	return domain.Directory{Path: dir, Equipment: equipment, Instruments: instruments, Vocabulary: vocabulary}, err
}

func readGroup(dir, folder, suffix string) ([]domain.File, error) {
	folderPath := filepath.Join(dir, folder)
	names, err := kernel.Names(folderPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("component: list %s: %w", folderPath, err)
	}
	return readAll(folderPath, sortedWithSuffix(names, suffix))
}

func readAll(folder string, names []string) ([]domain.File, error) {
	var files []domain.File
	for _, name := range names {
		file, err := readFile(filepath.Join(folder, name))
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func sortedWithSuffix(names []string, suffix string) []string {
	matching := slices.DeleteFunc(slices.Clone(names), func(name string) bool { return !strings.HasSuffix(name, suffix) })
	slices.Sort(matching)
	return matching
}

func readFile(path string) (domain.File, error) {
	raw, err := kernel.Read(path)
	if err != nil {
		return domain.File{}, fmt.Errorf("component: read %s: %w", path, err)
	}
	return domain.File{Path: path, Raw: raw}, nil
}

func readVocabulary(dir string) (*domain.File, error) {
	path := filepath.Join(dir, vocabularyFile)
	raw, present, err := kernel.ReadIfPresent(path)
	if err != nil {
		return nil, fmt.Errorf("component: read %s: %w", path, err)
	}
	if !present {
		return nil, nil
	}
	return &domain.File{Path: path, Raw: raw}, nil
}
