package architecture

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// productionFile is one parsed non-test Go source file of the repository.
type productionFile struct {
	path    string // slash-separated, relative to the repository root
	pkgDir  string // owning package directory, "." for the root package
	source  *ast.File
	paths   []string          // every imported path, blank and dot imports included
	imports map[string]string // local name -> import path, for selector resolution
}

var (
	loadOnce  sync.Once
	loadedSet []productionFile
	loadErr   error
)

// productionFiles parses every production Go file once and shares the result
// between gates, so each gate reads one snapshot of the repository.
func productionFiles(t *testing.T) []productionFile {
	t.Helper()
	loadOnce.Do(func() { loadedSet, loadErr = parseProduction("../..") })
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	return loadedSet
}

func parseProduction(root string) ([]productionFile, error) {
	var files []productionFile
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect %s: %w", path, err)
		}
		if entry.IsDir() {
			return skipDirectory(entry.Name())
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parseProductionFile(root, path)
		if err != nil {
			return err
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	return files, nil
}

func skipDirectory(name string) error {
	if name == ".git" || name == ".enola" || name == "testdata" {
		return fs.SkipDir
	}
	return nil
}

func parseProductionFile(root, path string) (productionFile, error) {
	source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return productionFile{}, fmt.Errorf("parse %s: %w", path, err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return productionFile{}, fmt.Errorf("locate %s: %w", path, err)
	}
	relative = filepath.ToSlash(relative)
	return productionFile{
		path:    relative,
		pkgDir:  packageDirectory(relative),
		source:  source,
		paths:   importPaths(source),
		imports: importNames(source),
	}, nil
}

func packageDirectory(relative string) string {
	directory := filepath.ToSlash(filepath.Dir(relative))
	if directory == "." && filepath.Base(relative) == "tools.go" {
		return "tools" // build-tagged development tools, outside runtime
	}
	return directory
}

func importPaths(source *ast.File) []string {
	var paths []string
	for _, spec := range source.Imports {
		if path, err := strconv.Unquote(spec.Path.Value); err == nil {
			paths = append(paths, path)
		}
	}
	return paths
}

func importNames(source *ast.File) map[string]string {
	names := map[string]string{}
	for _, spec := range source.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		names[localImportName(spec, path)] = path
	}
	return names
}

func localImportName(spec *ast.ImportSpec, path string) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	parts := strings.Split(path, "/")
	last := parts[len(parts)-1]
	if strings.HasPrefix(last, "v") && len(parts) > 1 && isMajorVersion(last) {
		return parts[len(parts)-2]
	}
	return last
}

func isMajorVersion(element string) bool {
	_, err := strconv.Atoi(strings.TrimPrefix(element, "v"))
	return err == nil
}

// repositoryRoot opens the repository root for tests that read non-Go files.
func repositoryRoot(t *testing.T) *os.Root {
	t.Helper()
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	return root
}
