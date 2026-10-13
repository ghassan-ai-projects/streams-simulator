package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil { //nolint:gosec // G703: the path is under t.TempDir and the name is a test constant
		t.Fatal(err)
	}
}

func builtInDir() string { return filepath.Join(testsupport.RepositoryRoot(), "components") }

func TestLoadReadsTheBuiltInLibrary(t *testing.T) {
	t.Parallel()
	library, err := Load([]string{builtInDir()})
	if err != nil {
		t.Fatal(err)
	}
	equipment, instruments := library.Counts()
	if equipment < 2 || instruments < 3 {
		t.Fatalf("library = %d equipment and %d instruments, want at least 2 and 3", equipment, instruments)
	}
}

func TestLoadTakesOnlyTheFilesOfTheLibraryLayout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "equipment", "notes.txt"), "not a component")
	write(t, filepath.Join(dir, "equipment", "kiln.json"), "{")
	write(t, filepath.Join(dir, "instruments", "other.component.json"), "{")
	write(t, filepath.Join(dir, "compositions", "plant.composition.json"), "{")
	library, err := Load([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if equipment, instruments := library.Counts(); equipment != 0 || instruments != 0 {
		t.Fatalf("library = %d/%d, want the stray files ignored", equipment, instruments)
	}
}

func TestLoadAcceptsADirectoryWithOnlyAVocabulary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "vocabulary.json"), `{"schema_version": "0.1", "unit_families": {"mass": {"base": "kg", "units": {"kg": {"scale": 1}}}}}`)
	if _, err := Load([]string{dir}); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRefusesADirectoryThatDoesNotExist(t *testing.T) {
	t.Parallel()
	_, err := Load([]string{filepath.Join(t.TempDir(), "absent")})
	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "component: list") {
		t.Fatalf("error = %v, want a not-exist error naming the listing", err)
	}
}

func TestLoadRefusesAFileItCannotRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	locked := filepath.Join(dir, "equipment", "kiln.component.json")
	write(t, locked, "{}")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	_, err := Load([]string{dir})
	if !errors.Is(err, fs.ErrPermission) || !strings.Contains(err.Error(), "component: read") {
		t.Fatalf("error = %v, want a permission error naming the read", err)
	}
}

func TestLoadRefusesAVocabularyItCannotRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	locked := filepath.Join(dir, "vocabulary.json")
	write(t, locked, "{}")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{dir}); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error = %v, want a permission error", err)
	}
}

func TestLoadRefusesAListingItCannotRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	locked := filepath.Join(dir, "instruments")
	write(t, filepath.Join(locked, "x.instrument.json"), "{}")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) }) //nolint:gosec // G302: restoring the owner-only directory so the temporary tree can be removed
	if _, err := Load([]string{dir}); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error = %v, want a permission error", err)
	}
}

func TestLoadRefusesTheReservedNamespaceInACustomerDirectory(t *testing.T) {
	t.Parallel()
	customer := t.TempDir()
	raw, err := os.ReadFile(filepath.Join(builtInDir(), "equipment", "ahu.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(customer, "equipment", "ahu.component.json"), string(raw))
	_, err = Load([]string{builtInDir(), customer})
	if err == nil || !strings.Contains(err.Error(), `reserved "streamsim/" namespace`) {
		t.Fatalf("error = %v, want the reserved namespace refused", err)
	}
}
