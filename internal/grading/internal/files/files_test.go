package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestReadReturnsTheBytesOfAPublishedFileAndNamesAMissingOne(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := Read(path)
	if err != nil || string(raw) != "{}\n" {
		t.Fatalf("read = %q (%v)", raw, err)
	}
	absent := filepath.Join(filepath.Dir(path), "absent.jsonl")
	_, err = Read(absent)
	var pathFailure *fs.PathError
	if !errors.Is(err, fs.ErrNotExist) || !errors.As(err, &pathFailure) || pathFailure.Path != absent {
		t.Fatalf("a missing file: err = %v, want a path failure naming %s", err, absent)
	}
}
