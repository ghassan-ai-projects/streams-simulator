package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestADiffSideIsADomainFileOrARunFolderReadFromRunJSONOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	domainPath := filepath.Join(root, "p.domain.json")
	write(t, domainPath, "{}")
	write(t, filepath.Join(root, "runs", "r-1", "run.json"), `{"domain":{}}`)
	write(t, filepath.Join(root, "runs", "r-1", "sealed", "journal.jsonl"), "secret")
	file, err := ReadDiffInput(domainPath)
	if err != nil || file.Run || string(file.Raw) != "{}" || file.Label != domainPath {
		t.Fatalf("ReadDiffInput(file) = %+v, %v, want the file's bytes under its path", file, err)
	}
	folder := filepath.Join(root, "runs", "r-1")
	run, err := ReadDiffInput(folder)
	if err != nil || !run.Run || string(run.Raw) != `{"domain":{}}` || run.Label != folder {
		t.Fatalf("ReadDiffInput(folder) = %+v, %v, want run.json of the folder", run, err)
	}
}

func TestADiffSideRefusesWhatIsNotThere(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o750); err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"a missing path": filepath.Join(root, "none.domain.json"), "a folder without run.json": filepath.Join(root, "empty")} {
		if _, err := ReadDiffInput(path); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: %v, want a missing file", name, err)
		}
	}
}

func TestDiffReadsBothSidesAndComparesThem(t *testing.T) {
	t.Parallel()
	pond := testsupport.Domain("aquaculture-pond")
	got, err := Diff(pond, pond)
	if err != nil || !got.Identity.SameDigest || got.A.Source.Label != pond {
		t.Fatalf("Diff = %+v, %v, want the pond equal to itself", got, err)
	}
	if _, err := Diff(pond, filepath.Join(t.TempDir(), "none")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Diff = %v, want the second side missing", err)
	}
	if _, err := Diff(filepath.Join(t.TempDir(), "none"), pond); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Diff = %v, want the first side missing", err)
	}
}

func TestADiffSideLargerThanTheLimitIsRefused(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "big.domain.json")
	if err := os.WriteFile(path, make([]byte, maxRunFile+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadDiffInput(path); !errors.Is(err, kernel.ErrTooLarge) {
		t.Fatalf("ReadDiffInput = %v, want ErrTooLarge", err)
	}
}
