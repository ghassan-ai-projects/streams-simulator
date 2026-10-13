// Package files is the file-system edge of the component module: it reads the
// embedded built-in library and the customer directories of a component
// library and hands the bytes to the pure layer.
package files

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
)

const vocabularyFile = "vocabulary.json"

//go:embed builtin/vocabulary.json builtin/equipment/*.component.json builtin/instruments/*.instrument.json
var builtInLibrary embed.FS

func Load(customerDirs []string) (*domain.Library, error) {
	return loadFrom(embeddedTree{files: builtInLibrary, root: "builtin"}, customerDirs)
}

func loadFrom(builtIn embeddedTree, customerDirs []string) (*domain.Library, error) {
	first, err := readDirectory(builtIn, ".")
	if err != nil {
		return nil, err
	}
	directories := []domain.Directory{first}
	for _, dir := range domain.CustomerDirectories(customerDirs, nil) {
		directory, err := readDirectory(diskTree{}, dir)
		if err != nil {
			return nil, err
		}
		directories = append(directories, directory)
	}
	return domain.LoadLibrary(directories)
}

func readDirectory(source tree, dir string) (domain.Directory, error) {
	if _, err := source.names(dir); err != nil {
		return domain.Directory{}, fmt.Errorf("component: list %s: %w", source.label(dir), err)
	}
	equipment, err := readGroup(source, dir, "equipment", ".component.json")
	if err != nil {
		return domain.Directory{}, err
	}
	instruments, err := readGroup(source, dir, "instruments", ".instrument.json")
	if err != nil {
		return domain.Directory{}, err
	}
	vocabulary, err := readVocabulary(source, dir)
	return domain.Directory{Path: source.label(dir), Equipment: equipment, Instruments: instruments, Vocabulary: vocabulary}, err
}

func readGroup(source tree, dir, folder, suffix string) ([]domain.File, error) {
	folderPath := source.join(dir, folder)
	names, err := source.names(folderPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("component: list %s: %w", source.label(folderPath), err)
	}
	return readAll(source, folderPath, sortedWithSuffix(names, suffix))
}

func readAll(source tree, folder string, names []string) ([]domain.File, error) {
	var files []domain.File
	for _, name := range names {
		file, err := readFile(source, source.join(folder, name))
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

func readFile(source tree, path string) (domain.File, error) {
	raw, err := source.read(path)
	if err != nil {
		return domain.File{}, fmt.Errorf("component: read %s: %w", source.label(path), err)
	}
	return domain.File{Path: source.label(path), Raw: raw}, nil
}

func readVocabulary(source tree, dir string) (*domain.File, error) {
	path := source.join(dir, vocabularyFile)
	raw, present, err := source.readIfPresent(path)
	if err != nil {
		return nil, fmt.Errorf("component: read %s: %w", source.label(path), err)
	}
	if !present {
		return nil, nil
	}
	return &domain.File{Path: source.label(path), Raw: raw}, nil
}

func CustomerDirectories(flags []string, environment string) []string {
	return domain.CustomerDirectories(flags, filepath.SplitList(environment))
}
