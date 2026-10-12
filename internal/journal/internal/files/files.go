// Package files is the file-system edge of the journal module: the two
// append-only JSON Lines files of a run, written through a buffer, flushed at
// command boundaries and synced when the run ends.
package files

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
)

// Open creates the observer journal at dir/journal.jsonl and the director
// journal at dir/sealed/journal.jsonl (and their directories) and returns the
// journal writing to them. retain also keeps the bytes in memory.
func Open(dir string, retain bool) (*domain.Journal, error) {
	observer, err := openSink(filepath.Join(dir, domain.ObserverFile))
	if err != nil {
		return nil, err
	}
	director, err := openSink(filepath.Join(dir, domain.DirectorFile))
	if err != nil {
		_ = observer.Close()
		return nil, err
	}
	return domain.New(domain.NewStream(domain.StreamObserver, observer, retain),
		domain.NewStream(domain.StreamDirector, director, retain)), nil
}

// sink is a buffered append-only file.
type sink struct {
	file   *os.File
	writer *bufio.Writer
}

func openSink(path string) (*sink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("journal: directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("journal: open %s: %w", path, err)
	}
	return &sink{file: file, writer: bufio.NewWriter(file)}, nil
}

// Write buffers one complete line.
func (s *sink) Write(line []byte) error {
	if _, err := s.writer.Write(line); err != nil {
		return fmt.Errorf("write %s: %w", s.file.Name(), err)
	}
	return nil
}

// Flush pushes the buffer to the file descriptor without syncing.
func (s *sink) Flush() error {
	if err := s.writer.Flush(); err != nil {
		return fmt.Errorf("flush %s: %w", s.file.Name(), err)
	}
	return nil
}

// Close flushes, syncs and closes the file. The file is closed even when the
// flush or sync fails.
func (s *sink) Close() error {
	flushErr := s.Flush()
	syncErr := s.file.Sync()
	closeErr := s.file.Close()
	for _, err := range []error{flushErr, syncErr, closeErr} {
		if err != nil {
			return fmt.Errorf("close %s: %w", s.file.Name(), err)
		}
	}
	return nil
}
