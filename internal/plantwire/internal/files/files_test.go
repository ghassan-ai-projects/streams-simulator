package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func baseNames(sources []domain.KindSource) []string {
	names := make([]string, len(sources))
	for i, source := range sources {
		names[i] = filepath.Base(source.Path) + "=" + string(source.Raw)
	}
	return names
}

func TestDirectoriesReadOnlyKindFilesEachInNameOrder(t *testing.T) {
	t.Parallel()
	first, second := t.TempDir(), t.TempDir()
	writeFile(t, first, "b.kind.json", "b")
	writeFile(t, first, "a.kind.json", "a")
	writeFile(t, first, "notes.json", "not a kind file")
	writeFile(t, second, "c.kind.json", "c")
	if err := os.Mkdir(filepath.Join(first, "folder.kind.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	sources, err := Directories([]string{first, second})
	if got, want := strings.Join(baseNames(sources), " "), "a.kind.json=a b.kind.json=b c.kind.json=c"; err != nil || got != want {
		t.Fatalf("Directories = %q, %v, want %q", got, err, want)
	}
}

func TestDirectoriesWithoutAnyDirectoryGiveNoSource(t *testing.T) {
	t.Parallel()
	if sources, err := Directories(nil); err != nil || len(sources) != 0 {
		t.Fatalf("Directories(nil) = %v, %v, want nothing", sources, err)
	}
}

func TestDirectoriesNameTheDirectoryTheyCannotList(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "kinds")
	_, err := Directories([]string{missing})
	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "list "+missing) {
		t.Fatalf("Directories = %v, want a missing directory named in the error", err)
	}
}

func TestPathsReadTheGivenFilesInOrderAndNameOneThatIsMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	second, first := writeFile(t, dir, "z.kind.json", "z"), writeFile(t, dir, "y.kind.json", "y")
	sources, err := Paths([]string{second, first})
	if got, want := strings.Join(baseNames(sources), " "), "z.kind.json=z y.kind.json=y"; err != nil || got != want {
		t.Fatalf("Paths = %q, %v, want %q", got, err, want)
	}
	missing := filepath.Join(dir, "absent.kind.json")
	if _, err := Paths([]string{first, missing}); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "read "+missing) {
		t.Fatalf("Paths = %v, want the missing file named in the error", err)
	}
}
