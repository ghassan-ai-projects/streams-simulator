// Package sockets is the kernel's unix-socket edge: sockets published owner-only, and dialed only when owner-only.
package sockets

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

var (
	ErrNotPrivate = errors.New("the socket is not owner-only")
	ErrInUse      = errors.New("the socket path is in use")
)

func Listen(ctx context.Context, path string) (*net.UnixListener, error) {
	if err := clearStale(ctx, path); err != nil {
		return nil, err
	}
	binding := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".binding")
	listener, err := bindAt(ctx, binding)
	if err != nil {
		return nil, err
	}
	if err := publish(binding, path); err != nil {
		_ = listener.Close()
		return nil, err
	}
	return listener, nil
}

func ListenPrivate(ctx context.Context, path string) (*net.UnixListener, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if err := Owned(dir, os.ModeDir); err != nil {
		return nil, err
	}
	return Listen(ctx, path)
}

func DialPrivate(ctx context.Context, path string) (net.Conn, error) {
	if err := Owned(filepath.Dir(path), os.ModeDir); err != nil {
		return nil, err
	}
	if err := Owned(path, os.ModeSocket); err != nil {
		return nil, err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", path, err)
	}
	return conn, nil
}

func Owned(target string, kind os.FileMode) error {
	info, err := os.Lstat(target)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	owner, ok := info.Sys().(*syscall.Stat_t)
	if info.Mode().Perm()&0o077 != 0 || info.Mode().Type() != kind || !ok || int(owner.Uid) != os.Geteuid() {
		return fmt.Errorf("%w: %s (mode %v)", ErrNotPrivate, target, info.Mode())
	}
	return nil
}

func clearStale(ctx context.Context, path string) error {
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("%w", err)
	case info.Mode().Type() != os.ModeSocket:
		return fmt.Errorf("%w: %s is not a socket", ErrInUse, path)
	case listening(ctx, path):
		return fmt.Errorf("%w: something already listens on %s", ErrInUse, path)
	}
	return removed(path)
}

func listening(ctx context.Context, path string) bool {
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func removed(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func bindAt(ctx context.Context, binding string) (*net.UnixListener, error) {
	_ = os.Remove(binding)
	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", binding)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", binding, err)
	}
	unix := listener.(*net.UnixListener)
	unix.SetUnlinkOnClose(false)
	return unix, nil
}

func publish(binding, path string) error {
	if err := os.Chmod(binding, 0o600); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := os.Rename(binding, path); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
