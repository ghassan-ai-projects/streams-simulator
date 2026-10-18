package component

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func builtInLibrary(t *testing.T) *Library {
	t.Helper()
	library, err := NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	return library
}

func TestTheBuiltInLibraryListsItsEquipmentAndInstruments(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	equipment, instruments := library.Counts()
	wantEquipment, wantInstruments := builtInFileCounts(t)
	if equipment != wantEquipment || instruments != wantInstruments || len(library.List()) != wantEquipment+wantInstruments {
		t.Fatalf("library = %d equipment, %d instruments, %d entries; want %d, %d, %d", equipment, instruments, len(library.List()),
			wantEquipment, wantInstruments, wantEquipment+wantInstruments)
	}
}

func builtInFileCounts(t *testing.T) (equipment, instruments int) {
	t.Helper()
	root := testsupport.ComponentsDir()
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
	raw, err := os.ReadFile(testsupport.PilotComposition())
	if err != nil {
		t.Fatal(err)
	}
	composed, err := Compose(builtInLibrary(t), raw, "chiller-pilot.composition.json")
	if err != nil || len(composed.Bytes) == 0 || composed.Domain.Spec.ID != "chiller-pilot" {
		t.Fatalf("Compose = %+v, %v", composed, err)
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
	if _, err := NewLibrary([]string{filepath.Join(t.TempDir(), "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want a not-exist error", err)
	}
}

func TestTheBuiltInLibraryLoadsWithoutAnyDirectoryAndIgnoresEmptyEntries(t *testing.T) {
	t.Parallel()
	for name, dirs := range map[string][]string{"none": nil, "empty entries": {"", " ", ""}} {
		library, err := NewLibrary(dirs)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		wantEquipment, _ := filepath.Glob(filepath.Join(testsupport.ComponentsDir(), "equipment", "*.component.json"))
		wantInstruments, _ := filepath.Glob(filepath.Join(testsupport.ComponentsDir(), "instruments", "*.instrument.json"))
		if equipment, instruments := library.Counts(); equipment != len(wantEquipment) || instruments != len(wantInstruments) {
			t.Errorf("%s: library = %d equipment, %d instruments, want %d and %d", name, equipment, instruments, len(wantEquipment), len(wantInstruments))
		}
	}
}

func TestACustomerDirectoryCannotUseTheStreamsimNamespace(t *testing.T) {
	t.Parallel()
	customer := t.TempDir()
	raw, err := os.ReadFile(filepath.Join(testsupport.ComponentsDir(), "equipment", "ahu.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(customer, "equipment")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, folder, "ahu.component.json", raw)
	_, err = NewLibrary([]string{customer})
	if err == nil || !strings.Contains(err.Error(), `reserved "streamsim/" namespace`) {
		t.Fatalf("error = %v, want the reserved namespace refused", err)
	}
}

func TestCustomerDirectoriesAreTheFlagsThenTheEnvironment(t *testing.T) {
	t.Parallel()
	if got := CustomerDirectories([]string{"/a"}, "/b:"); len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
		t.Fatalf("directories = %v", got)
	}
}
