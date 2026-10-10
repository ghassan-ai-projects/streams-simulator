package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

func TestWriteThenReadRoundTripsOwnerOnly(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "doc.json")
	if err := Write(path, []byte("{}")); err != nil {
		t.Fatal(err)
	}
	raw, err := Read(path)
	if err != nil || string(raw) != "{}" {
		t.Fatalf("read = %q (%v)", raw, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v (%v), want 0600", info.Mode().Perm(), err)
	}
}

func TestReadAndWriteKeepTheOperatingSystemsErrorText(t *testing.T) {
	t.Parallel()
	absent := filepath.Join(t.TempDir(), "absent", "doc")
	_, readErr := os.ReadFile(absent)
	if _, err := Read(absent); err == nil || err.Error() != readErr.Error() {
		t.Fatalf("read error = %v, want %v", err, readErr)
	}
	if err := Write(absent, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a write under a missing directory: err = %v, want ErrNotExist", err)
	}
}

func TestEnsureDirCreatesParentsOwnerOnly(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "a", "b")
	if err := EnsureDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("dir = %v (%v)", info, err)
	}
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDir(filepath.Join(blocker, "sub")); !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("a directory under a regular file: err = %v, want ENOTDIR", err)
	}
}

func TestFileNamesListsRegularEntriesInNameOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, name := range []string{"b.json", "a.json", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	names, err := FileNames(dir)
	if err != nil || !slices.Equal(names, []string{".hidden", "a.json", "b.json"}) {
		t.Fatalf("names = %v (%v)", names, err)
	}
	if _, err := FileNames(filepath.Join(dir, "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing directory: err = %v, want ErrNotExist", err)
	}
}
