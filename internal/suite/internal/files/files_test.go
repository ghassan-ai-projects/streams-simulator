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
