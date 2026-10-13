package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAnAtomicWriteReplacesTheFileOwnerOnlyAndLeavesNoTemporary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "run.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil { //nolint:gosec // the test starts from a readable file on purpose
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v, %v", info, err)
	}
	if raw, _ := Read(path); string(raw) != "new" {
		t.Fatalf("contents %q", raw)
	}
	if names, _ := Names(dir); !slices.Equal(names, []string{"run.json"}) {
		t.Fatalf("files left %v", names)
	}
}

func TestAFailedAtomicWriteKeepsTheOldFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "target")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	var pathErr *fs.PathError
	if err := WriteAtomic(path, []byte("x")); !errors.As(err, &pathErr) || pathErr.Path != path {
		t.Fatalf("an atomic write over a directory = %v", err)
	}
	if names, _ := Names(dir); len(names) != 0 {
		t.Fatalf("a temporary was left: %v", names)
	}
	if err := WriteAtomic(filepath.Join(dir, "missing", "f"), nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a write into a missing directory = %v", err)
	}
}

func TestReadIfPresentTellsAMissingFileFromAFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if raw, present, err := ReadIfPresent(filepath.Join(dir, "none")); raw != nil || present || err != nil {
		t.Fatalf("missing = %v, %v, %v", raw, present, err)
	}
	var pathErr *fs.PathError
	if _, present, err := ReadIfPresent(dir); present || !errors.As(err, &pathErr) {
		t.Fatalf("a directory = %v, %v", present, err)
	}
	if _, err := Names(filepath.Join(dir, "none")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("names of a missing directory = %v", err)
	}
}

func TestALogTruncatesOrKeepsAndIsOwnerOnly(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sealed", "log.jsonl")
	for _, keep := range []bool{false, true, false} {
		log, err := OpenLog(path, keep)
		if err != nil {
			t.Fatal(err)
		}
		if err := errors.Join(log.Write([]byte("a\n")), log.Flush(), log.Close()); err != nil {
			t.Fatal(err)
		}
		if raw, _ := Read(path); (keep && string(raw) != "a\na\n") || (!keep && string(raw) != "a\n") {
			t.Fatalf("keep %v: %q", keep, raw)
		}
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("log mode %v", info.Mode())
	}
	if dir, _ := os.Stat(filepath.Dir(path)); dir.Mode().Perm() != 0o700 {
		t.Fatalf("log directory mode %v", dir.Mode())
	}
}

func TestALogClosesItsFileEvenWhenItsFileWasClosedUnderIt(t *testing.T) {
	t.Parallel()
	log, err := OpenLog(filepath.Join(t.TempDir(), "log.jsonl"), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := log.file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := log.Write([]byte("x\n")); err != nil {
		t.Fatalf("a buffered write cannot fail yet: %v", err)
	}
	if err := log.Flush(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("flush over a closed file = %v", err)
	}
	if err := log.Close(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("close over a closed file = %v", err)
	}
}

func TestALogUnderAFileCannotOpen(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var pathErr *fs.PathError
	if _, err := OpenLog(filepath.Join(blocker, "log"), false); !errors.As(err, &pathErr) {
		t.Fatalf("a log under a file = %v", err)
	}
}
