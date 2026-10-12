package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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
