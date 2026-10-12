package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVerdictThenReadTraceRoundTripsOwnerOnly(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "verdict.json")
	if err := WriteVerdict(path, []byte("{}")); err != nil {
		t.Fatal(err)
	}
	raw, err := ReadTrace(path)
	if err != nil || string(raw) != "{}" {
		t.Fatalf("read = %q (%v)", raw, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v (%v), want 0600", info.Mode().Perm(), err)
	}
}

func TestReadTraceAndWriteVerdictNameThePathTheyCannotUse(t *testing.T) {
	t.Parallel()
	absent := filepath.Join(t.TempDir(), "absent", "file")
	var pathFailure *fs.PathError
	if _, err := ReadTrace(absent); !errors.As(err, &pathFailure) || pathFailure.Path != absent {
		t.Fatalf("a missing trace: err = %v", err)
	}
	if err := WriteVerdict(absent, nil); !errors.As(err, &pathFailure) || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a verdict under a missing directory: err = %v", err)
	}
}
