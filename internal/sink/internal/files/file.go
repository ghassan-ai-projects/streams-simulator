// Package files is the file-system edge of the sink module: a buffered,
// byte-reproducible line sink backed by one file.
package files

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// File writes lines to a path, buffered; Close flushes.
type File struct {
	path string
	f    *os.File
	w    *bufio.Writer
	mu   sync.Mutex
}

// New opens the sink at path (created if missing).
func New(path string) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil && filepath.Dir(path) != "." {
		return nil, fmt.Errorf("sink: mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("sink: create %s: %w", path, err)
	}
	return &File{path: path, f: f, w: bufio.NewWriter(f)}, nil
}

// Write appends a line.
func (s *File) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.w.Write(line); err != nil {
		return fmt.Errorf("sink: write %s: %w", s.path, err)
	}
	if err := s.w.WriteByte('\n'); err != nil {
		return fmt.Errorf("sink: write %s: %w", s.path, err)
	}
	return nil
}

// Flush pushes buffered bytes to the file descriptor without syncing. The
// delivery ledger and the trace stay consistent at command boundaries: a
// crash after a boundary leaves both recoverable from the page cache; End
// syncs them to disk.
func (s *File) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.w.Flush(); err != nil {
		return fmt.Errorf("sink: flush %s: %w", s.path, err)
	}
	return nil
}

// Close flushes and returns the file contents.
func (s *File) Close() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.finishFile(); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("sink: read back %s: %w", s.path, err)
	}
	return raw, nil
}

func (s *File) finishFile() error {
	if err := s.w.Flush(); err != nil {
		return fmt.Errorf("sink: flush %s: %w", s.path, err)
	}
	if err := s.f.Sync(); err != nil {
		return fmt.Errorf("sink: sync %s: %w", s.path, err)
	}
	if err := s.f.Close(); err != nil {
		return fmt.Errorf("sink: close %s: %w", s.path, err)
	}
	return nil
}
