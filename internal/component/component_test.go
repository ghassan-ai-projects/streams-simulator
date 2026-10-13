package component

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func builtInLibrary(t *testing.T) *Library {
	t.Helper()
	library, err := NewLibraryFromDirs([]string{filepath.Join(testsupport.RepositoryRoot(), "components")})
	if err != nil {
		t.Fatal(err)
	}
	return library
}

func builtInFileCounts(t *testing.T) (equipment, instruments int) {
	t.Helper()
	root := filepath.Join(testsupport.RepositoryRoot(), "components")
	equipmentFiles, err := filepath.Glob(filepath.Join(root, "equipment", "*.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	instrumentFiles, err := filepath.Glob(filepath.Join(root, "instruments", "*.instrument.json"))
	if err != nil {
		t.Fatal(err)
	}
	return len(equipmentFiles), len(instrumentFiles)
}

func TestTheBuiltInLibraryListsItsEquipmentAndInstruments(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	equipment, instruments := library.Counts()
	wantEquipment, wantInstruments := builtInFileCounts(t)
	if equipment != wantEquipment || instruments != wantInstruments || len(library.List()) != wantEquipment+wantInstruments {
		t.Fatalf("library = %d equipment, %d instruments, %d entries; want %d, %d files", equipment, instruments, len(library.List()), wantEquipment, wantInstruments)
	}
}

func TestTheBuiltInLibraryDescribesAnEntryByItsReferenceOrItsHighestMajor(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	for _, ref := range []string{"streamsim/chiller@1", "streamsim/chiller"} {
		detail, err := library.Describe(ref)
		if err != nil || detail.Entry.ID != "streamsim/chiller" || detail.Document["id"] != "streamsim/chiller" {
			t.Errorf("Describe(%s) = %+v (%v), want the chiller", ref, detail, err)
		}
	}
}

func TestComposeWritesTheBuiltInPilotAsADomain(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "components", "compositions", "chiller-pilot.composition.json"))
	if err != nil {
		t.Fatal(err)
	}
	composed, err := Compose(builtInLibrary(t), raw, "chiller-pilot.composition.json")
	if err != nil || len(composed) == 0 {
		t.Fatalf("Compose = %d bytes, %v", len(composed), err)
	}
}

func TestALibraryThatWasNeverBuiltRefusesEveryOperation(t *testing.T) {
	t.Parallel()
	for name, library := range map[string]*Library{"nil": nil, "zero": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if entries := library.List(); entries != nil {
				t.Errorf("List = %v, want nil", entries)
			}
			if equipment, instruments := library.Counts(); equipment != 0 || instruments != 0 {
				t.Errorf("Counts = %d, %d, want 0, 0", equipment, instruments)
			}
			if _, err := library.Describe("streamsim/chiller"); !errors.Is(err, ErrNoLibrary) {
				t.Errorf("Describe error = %v, want ErrNoLibrary", err)
			}
			if _, err := Compose(library, []byte("{}"), "x"); !errors.Is(err, ErrNoLibrary) {
				t.Errorf("Compose error = %v, want ErrNoLibrary", err)
			}
		})
	}
}

func TestALibraryDirectoryThatDoesNotExistIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := NewLibraryFromDirs([]string{filepath.Join(t.TempDir(), "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want a not-exist error", err)
	}
}
