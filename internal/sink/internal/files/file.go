// Package files is the file-system edge of the sink module: a buffered,
// byte-reproducible line sink backed by one file.
package files

import (
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

type File struct {
	path string
	log  *kernel.Log
	mu   sync.Mutex
}

func New(path string) (*File, error) {
	log, err := kernel.NewLog(path, false)
	if err != nil {
		return nil, fmt.Errorf("sink: create %s: %w", path, err)
	}
	return &File{path: path, log: log}, nil
}

func (s *File) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.log.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("sink: %w", err)
	}
	return nil
}

func (s *File) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.log.Flush(); err != nil {
		return fmt.Errorf("sink: %w", err)
	}
	return nil
}

func (s *File) Close() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.log.Close(); err != nil {
		return nil, fmt.Errorf("sink: %w", err)
	}
	raw, err := kernel.Read(s.path)
	if err != nil {
		return nil, fmt.Errorf("sink: read back %s: %w", s.path, err)
	}
	return raw, nil
}
