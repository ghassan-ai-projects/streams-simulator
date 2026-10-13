package architecture

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var docsReference = regexp.MustCompile(`(^|[^A-Za-z_-])docs/|"docs"`)

var checkedHiddenDirectories = map[string]bool{".github": true, ".agents": true}

const docsBoundaryGate = "test/architecture/docs_boundary_test.go"

func outsideRepositoryContent(path string) bool {
	name := path[strings.LastIndex(path, "/")+1:]
	return path == "docs" || path == "bin" || name == "node_modules" || (strings.HasPrefix(name, ".") && name != "." && !checkedHiddenDirectories[name])
}

// TestNothingOutsideDocsDependsOnDocs keeps docs/ a working directory: code,
// tests, scripts, contracts and data never read from it, so it can leave git.
func TestNothingOutsideDocsDependsOnDocs(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	err := fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && outsideRepositoryContent(path) {
			return fs.SkipDir
		}
		if entry.IsDir() || strings.HasSuffix(path, ".md") || path == docsBoundaryGate {
			return nil
		}
		raw, err := root.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if docsReference.Match(raw) {
			t.Errorf("%s refers to docs/, which holds working documents only", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
