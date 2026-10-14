package architecture

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

const (
	customerWorkspaceMarker = "STREAMSIM_CUSTOMER_WORKSPACE"
	canvasGate              = "test/architecture/canvas_test.go"
)

var canvasPath = regexp.MustCompile(`["'` + "`" + `]\.{0,2}/?canvas/`)

// TestGoNeverReachesIntoTheCanvas keeps the one-way dependency: no Go file lives under canvas/
// and no Go source names a path under it, so the Go build never depends on the canvas build.
func TestGoNeverReachesIntoTheCanvas(t *testing.T) {
	t.Parallel()
	for _, file := range walkRepository(t) {
		switch {
		case strings.HasPrefix(file, "canvas/") && strings.HasSuffix(file, ".go"):
			t.Errorf("%s: Go code does not live in canvas/", file)
		case strings.HasSuffix(file, ".go") && file != canvasGate && canvasPath.Match(readRepositoryFile(t, file)):
			t.Errorf("%s: Go code names a path under canvas/; the bridge serves the built UI from --ui-dir", file)
		}
	}
}

// TestNoCustomerWorkspaceInsideTheRepository refuses any folder carrying the customer-workspace
// marker: customer plants live outside every repository (REVIEW I3, plant-wire §16).
func TestNoCustomerWorkspaceInsideTheRepository(t *testing.T) {
	t.Parallel()
	for _, file := range walkRepository(t) {
		if strings.HasSuffix(file, "/"+customerWorkspaceMarker) || file == customerWorkspaceMarker {
			t.Errorf("%s: a customer workspace is inside the repository; move it out", file)
		}
	}
}

func walkRepository(t *testing.T) []string {
	t.Helper()
	root := repositoryRoot(t)
	var files []string
	err := fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}
		if entry.IsDir() && isOutsideRepositoryCode(path) {
			return fs.SkipDir
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func readRepositoryFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := repositoryRoot(t).ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
