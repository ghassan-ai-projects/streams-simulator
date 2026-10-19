// Package files is the kernel's file-system edge: reads, owner-only directories and atomic owner-only writes.
package files

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Read(path string) ([]byte, error) {
	// #nosec G304 -- every caller reads a path its user or its own run named.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
}

var (
	ErrNotRegular = errors.New("kernel: not a regular file")
	ErrTooLarge   = errors.New("kernel: file exceeds its size limit")
	ErrSymlink    = errors.New("kernel: a path component is a symbolic link")
)

func ReadRegular(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", ErrNotRegular, path)
	}
	return readUpTo(path, limit)
}

func readUpTo(path string, limit int64) ([]byte, error) {
	// #nosec G304 -- every caller reads a path its user named, after it was found to be a regular file.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	raw, read := io.ReadAll(io.LimitReader(file, limit+1))
	if failed := errors.Join(read, file.Close()); failed != nil {
		return nil, fmt.Errorf("%w", failed)
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("%w: %s is larger than %d bytes", ErrTooLarge, path, limit)
	}
	return raw, nil
}

func OpenIfPresent(path string) (io.ReadCloser, bool, error) {
	// #nosec G304 -- every caller reads a path its user or its own run named.
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w", err)
	}
	return file, true, nil
}

func ReadIfPresent(path string) ([]byte, bool, error) {
	raw, err := Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return raw, err == nil, err
}

func Names(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func PrivateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func WriteAtomic(path string, data []byte) error {
	if err := replace(path, data, 0o600); err != nil {
		return &fs.PathError{Op: "write", Path: path, Err: err}
	}
	return nil
}

func Rename(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return &fs.PathError{Op: "rename", Path: from, Err: err}
	}
	return nil
}

func Replace(path string, data []byte) error {
	info, err := os.Stat(path)
	if err == nil {
		err = replace(path, data, info.Mode().Perm())
	}
	if err != nil {
		return &fs.PathError{Op: "replace", Path: path, Err: err}
	}
	return nil
}

func replace(path string, data []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := fill(temp, data, mode); err != nil {
		return errors.Join(err, os.Remove(temp.Name()))
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return errors.Join(fmt.Errorf("%w", err), os.Remove(temp.Name()))
	}
	return syncDir(filepath.Dir(path))
}

func fill(temp *os.File, data []byte, mode os.FileMode) error {
	_, written := temp.Write(data)
	synced := temp.Sync()
	closed := temp.Close()
	if err := cmp.Or(written, synced, closed, os.Chmod(temp.Name(), mode)); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func syncDir(dir string) error {
	handle, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	synced := handle.Sync()
	if err := cmp.Or(synced, handle.Close()); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func OpenRegular(path string) (io.ReadCloser, error) {
	if filepath.IsAbs(path) {
		return openLast(path)
	}
	return OpenRegularIn(".", path)
}

func openLast(path string) (io.ReadCloser, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", ErrNotRegular, path)
	}
	// #nosec G304 -- the path was found to be a regular file, not a link, just above.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return file, nil
}

func OpenRegularIn(base, rel string) (io.ReadCloser, error) {
	root, err := os.OpenRoot(base)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() { _ = root.Close() }()
	if err := inspectPath(root, rel); err != nil {
		return nil, err
	}
	file, err := root.Open(rel)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return file, nil
}

func inspectPath(root *os.Root, rel string) error {
	elements := strings.Split(filepath.ToSlash(filepath.Clean(rel)), "/")
	var info fs.FileInfo
	for i := range elements {
		var err error
		info, err = root.Lstat(strings.Join(elements[:i+1], "/"))
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s", ErrSymlink, strings.Join(elements[:i+1], "/"))
		}
	}
	return mustBeRegular(info, rel)
}

func mustBeRegular(info fs.FileInfo, rel string) error {
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%w: %s", ErrNotRegular, rel)
	}
	return nil
}
