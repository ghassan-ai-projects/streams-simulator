// Package testsupport holds the fixtures and builders shared by tests: where
// the shipped domains and adapters live, and short unix-socket paths. Only
// _test.go files import it; a production package never does.
package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

// RepositoryRoot is the directory holding the module's go.mod, found by
// walking up from the test's working directory.
func RepositoryRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("testsupport: working directory: " + err.Error())
	}
	for !hasGoMod(dir) {
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("testsupport: no go.mod above the working directory")
		}
		dir = parent
	}
	return dir
}

func hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

// DomainsDir is the directory of shipped domain specs.
func DomainsDir() string { return filepath.Join(RepositoryRoot(), "domains") }

// Domain is the path of the shipped domain spec with the given id.
func Domain(id string) string { return filepath.Join(DomainsDir(), id+".domain.json") }

// Example is the documented aquaculture-pond example domain.
func Example() string {
	return filepath.Join(RepositoryRoot(), "docs", "examples", "aquaculture-pond.domain.json")
}

// AdaptersDir is the directory of shipped adapters.
func AdaptersDir() string { return filepath.Join(RepositoryRoot(), "adapters") }

// Adapter is the path of the shipped adapter with the given id.
func Adapter(id string) string { return filepath.Join(AdaptersDir(), id+".adapter.json") }

// SocketPath returns a unix-socket path short enough for the platform's
// sun_path limit, in a directory removed when the test ends.
func SocketPath(tb testing.TB) string {
	tb.Helper()
	//nolint:usetesting // t.TempDir paths exceed the unix socket path limit on macOS
	dir, err := os.MkdirTemp("", "ss")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			tb.Error(err)
		}
	})
	return filepath.Join(dir, "d.sock")
}
