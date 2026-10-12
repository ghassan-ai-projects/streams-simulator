// Package files is the kernel's file-system edge: reads, owner-only directories and atomic owner-only writes.
package files

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Read(path string) ([]byte, error) {
	// #nosec G304 -- every caller reads a path its user or its own run named.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
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
