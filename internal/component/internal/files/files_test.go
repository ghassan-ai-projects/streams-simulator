package files

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
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

func builtInDir() string { return testsupport.ComponentsDir() }

func loadCustomer(dirs ...string) (int, int, error) {
	library, err := loadFrom(embeddedTree{files: fstest.MapFS{}, root: "."}, dirs)
	if err != nil {
		return 0, 0, err
	}
	equipment, instruments := library.Counts()
	return equipment, instruments, nil
}

func TestLoadReadsTheEmbeddedBuiltInLibrary(t *testing.T) {
	t.Parallel()
	library, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	equipment, instruments := library.Counts()
	wantEquipment, wantInstruments := builtInFileCounts(t)
	if equipment != wantEquipment || instruments != wantInstruments {
		t.Fatalf("library = %d equipment and %d instruments, want %d and %d", equipment, instruments, wantEquipment, wantInstruments)
	}
}

func builtInFileCounts(t *testing.T) (equipment, instruments int) {
	t.Helper()
	equipmentFiles, err := filepath.Glob(filepath.Join(builtInDir(), "equipment", "*.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	instrumentFiles, err := filepath.Glob(filepath.Join(builtInDir(), "instruments", "*.instrument.json"))
	if err != nil {
		t.Fatal(err)
	}
	return len(equipmentFiles), len(instrumentFiles)
}

func TestLoadTakesOnlyTheFilesOfTheLibraryLayout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "equipment", "notes.txt"), "not a component")
	write(t, filepath.Join(dir, "equipment", "kiln.json"), "{")
	write(t, filepath.Join(dir, "instruments", "other.component.json"), "{")
	write(t, filepath.Join(dir, "compositions", "plant.composition.json"), "{")
	equipment, instruments, err := loadCustomer(dir)
	if err != nil || equipment != 0 || instruments != 0 {
		t.Fatalf("library = %d/%d (%v), want the stray files ignored", equipment, instruments, err)
	}
}

func TestLoadAcceptsADirectoryWithOnlyAVocabulary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "vocabulary.json"), `{"schema_version": "0.1", "unit_families": {"mass": {"base": "kg", "units": {"kg": {"scale": 1}}}}}`)
	if _, _, err := loadCustomer(dir); err != nil {
		t.Fatal(err)
	}
}

func TestLoadIgnoresEmptyDirectoryEntries(t *testing.T) {
	t.Parallel()
	if _, _, err := loadCustomer("", "  ", ""); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRefusesADirectoryThatDoesNotExist(t *testing.T) {
	t.Parallel()
	_, _, err := loadCustomer(filepath.Join(t.TempDir(), "absent"))
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
	_, _, err := loadCustomer(dir)
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
	if _, _, err := loadCustomer(dir); !errors.Is(err, fs.ErrPermission) {
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
	if _, _, err := loadCustomer(dir); !errors.Is(err, fs.ErrPermission) {
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
	_, err = Load([]string{customer})
	if err == nil || !strings.Contains(err.Error(), `reserved "streamsim/" namespace`) {
		t.Fatalf("error = %v, want the reserved namespace refused", err)
	}
}

func TestLoadReadsTheBuiltInLibraryFromTheFileSystemItIsGiven(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(builtInDir(), "vocabulary.json"))
	if err != nil {
		t.Fatal(err)
	}
	library, err := loadFrom(embeddedTree{files: fstest.MapFS{"vocabulary.json": {Data: raw}}, root: "."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if equipment, instruments := library.Counts(); equipment != 0 || instruments != 0 {
		t.Fatalf("library = %d/%d, want only the vocabulary", equipment, instruments)
	}
}

func TestLoadNamesTheEmbeddedFileThatFailsToRead(t *testing.T) {
	t.Parallel()
	broken := fstest.MapFS{"equipment/kiln.component.json": {Data: []byte("{")}}
	if _, err := loadFrom(embeddedTree{files: broken, root: "."}, nil); err == nil || !strings.Contains(err.Error(), "builtin:equipment/kiln.component.json") {
		t.Fatalf("error = %v, want the embedded file named", err)
	}
}

var errUnreadableBuiltIn = errors.New("built-in library cannot be read")

type unreadableFS struct{}

func (unreadableFS) Open(string) (fs.File, error) { return nil, errUnreadableBuiltIn }

func TestLoadRefusesABuiltInLibraryItCannotList(t *testing.T) {
	t.Parallel()
	_, err := loadFrom(embeddedTree{files: unreadableFS{}, root: "."}, nil)
	if !errors.Is(err, errUnreadableBuiltIn) || !strings.Contains(err.Error(), "component: list builtin:") {
		t.Fatalf("error = %v, want the listing failure naming the built-in library", err)
	}
}

func TestCustomerDirectoriesSplitTheEnvironmentOnThePlatformListSeparator(t *testing.T) {
	t.Parallel()
	list := strings.Join([]string{"/acme/shared", "", "/acme/site"}, string(filepath.ListSeparator))
	got := CustomerDirectories([]string{"/acme/lib"}, list)
	if want := []string{"/acme/lib", "/acme/shared", "/acme/site"}; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("directories = %v, want %v", got, want)
	}
}

func TestTheBuiltInExamplesListAndAreFoundByID(t *testing.T) {
	t.Parallel()
	listed, err := ListExamples()
	if err != nil || len(listed) != 2 || listed[0].ID != "chiller-pilot" || listed[1].ID != "factory-utilities" {
		t.Fatalf("ListExamples = %+v, %v", listed, err)
	}
	found, err := FindExample("chiller-pilot")
	if err != nil || !strings.Contains(string(found), `"chiller-pilot"`) {
		t.Fatalf("FindExample = %.80s, %v", found, err)
	}
	if _, err := FindExample("nothing"); !errors.Is(err, domain.ErrUnknownExample) {
		t.Fatalf("unknown example = %v", err)
	}
	brokenFile := path.Join("examples", "x.composition.json")
	broken := embeddedTree{files: fstest.MapFS{brokenFile: {Data: []byte("{")}}, root: "."}
	if _, err := findExample(broken, "x"); err == nil || !strings.Contains(err.Error(), "x.composition.json") {
		t.Fatalf("a broken example file = %v", err)
	}
	if _, err := loadExamples(unlistable{broken}); !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("an unreadable example folder = %v", err)
	}
}

type unlistable struct{ embeddedTree }

func (unlistable) names(string) ([]string, error) { return nil, fs.ErrPermission }
