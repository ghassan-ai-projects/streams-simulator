package files

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteCreatesTheDirectoryAndTheFileOwnerOnly(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "a", "b")
	path, err := Write(dir, "suite.json", []byte("{}"))
	if err != nil || path != filepath.Join(dir, "suite.json") {
		t.Fatalf("path = %q, %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %v (%v), want 0600", info.Mode().Perm(), err)
	}
	if dirInfo, err := os.Stat(dir); err != nil || dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %v (%v), want 0700", dirInfo.Mode().Perm(), err)
	}
}

func TestWriteRefusesADirectoryUnderARegularFile(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(filepath.Join(blocker, "sub"), "suite.json", nil); !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("err = %v, want ENOTDIR", err)
	}
}

func TestReadReturnsTheFileAndRefusesAnAbsentOrOversizedOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path, err := Write(dir, "spec.json", []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if raw, err := Read(path); err != nil || string(raw) != "{}" {
		t.Fatalf("Read = %q, %v", raw, err)
	}
	if _, err := Read(filepath.Join(dir, "absent.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Read of an absent file = %v, want ErrNotExist", err)
	}
	big, err := Write(dir, "big.json", make([]byte, maxSpecBytes+1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Read(big); err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Read of an oversized file = %v, want a size refusal", err)
	}
}

func TestNamesListsTheFilesOfAFolderAndNothingForOneThatIsAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, name := range []string{"b.json", "a.json"} {
		if _, err := Write(dir, name, nil); err != nil {
			t.Fatal(err)
		}
	}
	if names, err := Names(dir); err != nil || len(names) != 2 {
		t.Fatalf("Names = %v, %v, want the two files", names, err)
	}
	if names, err := Names(filepath.Join(dir, "absent")); err != nil || len(names) != 0 {
		t.Fatalf("Names of an absent folder = %v, %v, want none", names, err)
	}
	if _, err := Names(filepath.Join(dir, "a.json")); !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("Names of a file = %v, want ENOTDIR", err)
	}
}
