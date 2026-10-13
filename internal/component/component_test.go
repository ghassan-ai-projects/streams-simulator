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

func TestTheBuiltInLibraryListsItsEquipmentAndInstruments(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	equipment, instruments := library.Counts()
	if equipment != 2 || instruments != 3 || len(library.List()) != 5 {
		t.Fatalf("library = %d equipment, %d instruments, %d entries; want 2, 3, 5", equipment, instruments, len(library.List()))
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
