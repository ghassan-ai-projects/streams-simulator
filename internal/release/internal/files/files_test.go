package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestWriteThenReadRoundTripsOwnerOnly(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.json")
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
	if err := Write(filepath.Join(t.TempDir(), "absent", "m.json"), nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a write under a missing directory: err = %v", err)
	}
}

func TestNamesListsRegularEntriesInNameOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, name := range []string{"b.json", "a.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	names, err := Names(dir)
	if err != nil || !slices.Equal(names, []string{"a.json", "b.json"}) {
		t.Fatalf("names = %v (%v)", names, err)
	}
	if _, err := Names(filepath.Join(dir, "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing directory: err = %v", err)
	}
}
