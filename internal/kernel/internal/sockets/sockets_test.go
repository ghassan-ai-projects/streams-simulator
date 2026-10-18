package sockets

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestASocketIsPublishedOwnerOnlyAndAStaleOneIsReplaced(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	stale, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	stale.(*net.UnixListener).SetUnlinkOnClose(false)
	_ = stale.Close()
	listener, err := ListenPrivate(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if err := Owned(path, os.ModeSocket); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(t.Context(), path); !errors.Is(err, ErrInUse) {
		t.Fatalf("a second listener on a live socket = %v", err)
	}
	conn, err := DialPrivate(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
}

func TestNeitherAFileNorASharedDirectoryIsUsed(t *testing.T) {
	t.Parallel()
	dir := filepath.Dir(testsupport.SocketPath(t))
	file := filepath.Join(dir, "plain")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(t.Context(), file); !errors.Is(err, ErrInUse) {
		t.Fatalf("listen on a regular file = %v", err)
	}
	shared := filepath.Join(dir, "shared")
	if err := os.Mkdir(shared, 0o700); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	if err := root.Chmod("shared", 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ListenPrivate(t.Context(), filepath.Join(shared, "s.sock")); !errors.Is(err, ErrNotPrivate) {
		t.Fatalf("listen in a shared directory = %v", err)
	}
	if _, err := DialPrivate(t.Context(), filepath.Join(shared, "s.sock")); !errors.Is(err, ErrNotPrivate) {
		t.Fatalf("dial in a shared directory = %v", err)
	}
}
