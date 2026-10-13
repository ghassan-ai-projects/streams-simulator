package files

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
)

type Log struct {
	file   *os.File
	writer *bufio.Writer
}

func OpenLog(path string, keep bool) (*Log, error) {
	if err := PrivateDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	mode := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if keep {
		mode = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	// #nosec G304 -- the run's own log file.
	file, err := os.OpenFile(path, mode, 0o600)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &Log{file: file, writer: bufio.NewWriter(file)}, nil
}

func (l *Log) Write(line []byte) error {
	if _, err := l.writer.Write(line); err != nil {
		return fmt.Errorf("write %s: %w", l.file.Name(), err)
	}
	return nil
}

func (l *Log) Flush() error {
	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("flush %s: %w", l.file.Name(), err)
	}
	return nil
}

func (l *Log) Close() error {
	flushed := l.writer.Flush()
	synced := l.file.Sync()
	closed := l.file.Close()
	if err := cmp.Or(flushed, synced, closed); err != nil {
		return fmt.Errorf("close %s: %w", l.file.Name(), err)
	}
	return nil
}
