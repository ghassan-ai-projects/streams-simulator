package kernel

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/files"

type Log struct {
	impl *files.Log
}

func NewLog(path string, keep bool) (*Log, error) {
	inner, err := files.OpenLog(path, keep)
	if err != nil {
		return nil, err
	}
	return &Log{impl: inner}, nil
}

func (l *Log) Write(line []byte) error {
	return l.impl.Write(line)
}

func (l *Log) Flush() error {
	return l.impl.Flush()
}

func (l *Log) Close() error {
	return l.impl.Close()
}
