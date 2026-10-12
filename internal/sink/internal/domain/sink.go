// Package domain holds the sink contract and the in-memory sink: lines go in,
// and Close returns the full byte stream the trace digest is computed from.
// It performs no I/O; file and HTTP sinks are edge packages of the module.
package domain

import "bytes"

type Sink interface {
	Write(line []byte) error
	Close() ([]byte, error)
}

type Inproc struct {
	buf bytes.Buffer
}

func (s *Inproc) Write(line []byte) error {
	s.buf.Write(line)
	s.buf.WriteByte('\n')
	return nil
}

func (s *Inproc) Close() ([]byte, error) {
	return s.buf.Bytes(), nil
}
