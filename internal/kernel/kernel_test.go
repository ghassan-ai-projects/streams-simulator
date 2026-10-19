package kernel

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheKernelFacadeReachesEveryEdge(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "out")
	if err := PrivateDir(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "a.json")
	if err := WriteAtomic(path, []byte("1")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	if err := root.Chmod("a.json", 0o640); err != nil {
		t.Fatal(err)
	}
	if err := Replace(path, []byte("2")); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o640 {
		t.Fatalf("a replaced file lost its mode: %v", info.Mode())
	}
	if raw, present, err := ReadIfPresent(path); string(raw) != "2" || !present || err != nil {
		t.Fatalf("read %q %v %v", raw, present, err)
	}
	if names, err := Names(dir); err != nil || len(names) != 1 {
		t.Fatalf("names %v %v", names, err)
	}
	if raw, err := ReadRegular(path, 1); string(raw) != "2" || err != nil {
		t.Fatalf("regular read %q %v", raw, err)
	}
	if _, err := ReadRegular(path, 0); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("a regular read over its limit = %v", err)
	}
	if _, err := ReadRegular(dir, 1); !errors.Is(err, ErrNotRegular) {
		t.Fatalf("a regular read of a directory = %v", err)
	}
	if err := Replace(filepath.Join(dir, "none"), nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("replace of a missing file = %v", err)
	}
	log, err := NewLog(filepath.Join(dir, "log"), true)
	if err != nil || errors.Join(log.Write([]byte("x")), log.Flush(), log.Close()) != nil {
		t.Fatalf("log %v", err)
	}
	if raw, err := Read(filepath.Join(dir, "log")); string(raw) != "x" || err != nil {
		t.Fatalf("log contents %q %v", raw, err)
	}
}

func TestTheKernelFacadeServesAPrivateSocket(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	listener, err := ListenPrivate(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if _, err := Listen(t.Context(), path); !errors.Is(err, ErrInUse) {
		t.Fatalf("a second listener = %v", err)
	}
	conn, err := DialPrivate(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	if _, err := DialPrivate(t.Context(), filepath.Join(t.TempDir(), "none.sock")); !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, ErrNotPrivate) {
		t.Fatalf("dial of a missing socket = %v", err)
	}
}

func TestAnInterruptContextEndsWithItsParent(t *testing.T) {
	t.Parallel()
	parent, cancel := context.WithCancel(t.Context())
	ctx, stop := UntilInterrupted(parent)
	defer stop()
	cancel()
	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("ctx = %v", ctx.Err())
	}
}

func TestTheKernelOpensARegularFileAndNamesTheLinkItRefuses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	regular, link := filepath.Join(dir, "a.json"), filepath.Join(dir, "link.json")
	if err := errors.Join(os.WriteFile(regular, []byte("1"), 0o600), os.Symlink(regular, link)); err != nil {
		t.Fatal(err)
	}
	file, err := OpenRegular(regular)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenRegular(link); !errors.Is(err, ErrNotRegular) {
		t.Fatalf("OpenRegular of a symbolic link = %v, want ErrNotRegular", err)
	}
	if ErrSymlink == nil {
		t.Fatal("the kernel names no symbolic link error")
	}
}

func TestRenameMovesAFileAndNamesTheSourceWhenItIsNotThere(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	from, to := filepath.Join(dir, "a.jsonl"), filepath.Join(dir, "b.jsonl")
	if err := os.WriteFile(from, []byte("1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Rename(from, to); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(to); err != nil || string(raw) != "1\n" {
		t.Fatalf("moved file = %q, %v, want its content", raw, err)
	}
	var pathErr *fs.PathError
	if err := Rename(from, to); !errors.Is(err, fs.ErrNotExist) || !errors.As(err, &pathErr) || pathErr.Op != "rename" || pathErr.Path != from {
		t.Fatalf("Rename of an absent file = %v, want a rename error naming %s", err, from)
	}
}
