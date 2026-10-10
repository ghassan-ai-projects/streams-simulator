// Package domain holds the sink contract and the in-memory sink: lines go in,
// and Close returns the full byte stream the trace digest is computed from.
// It performs no I/O; file and HTTP sinks are edge packages of the module.
package domain

import "bytes"

// Sink receives rendered lines and can produce the full byte stream at
// Close, which is what the trace digest is computed from.
type Sink interface {
	Write(line []byte) error
	Close() ([]byte, error)
}

// Inproc buffers everything in memory.
type Inproc struct {
	buf bytes.Buffer
}

// Write appends a line.
func (s *Inproc) Write(line []byte) error {
	s.buf.Write(line)
	s.buf.WriteByte('\n')
	return nil
}

// Close returns the buffer.
func (s *Inproc) Close() ([]byte, error) {
	return s.buf.Bytes(), nil
}
