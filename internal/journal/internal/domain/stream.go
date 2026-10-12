package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sync"
)

// Sink receives complete journal lines (each ends in a newline). The file
// edge implements it; a stream without a sink keeps only its digest.
type Sink interface {
	Write(line []byte) error
	Flush() error
	// Close flushes, syncs and closes the destination.
	Close() error
}

// Stream is one append-only stream of records. It numbers records, writes each
// as one line and keeps a running SHA-256 of the lines replay can reproduce.
// A failure to encode or write is remembered and returned by Flush and Close:
// a record is never dropped silently.
type Stream struct {
	mu     sync.Mutex
	name   string
	sink   Sink
	hash   hash.Hash
	keep   *bytes.Buffer
	seq    int64
	atNS   int64
	err    error
	closed bool
}

// NewStream starts a stream. A nil sink keeps the digest only; retain also
// keeps every line in memory (replay compares them).
func NewStream(name string, sink Sink, retain bool) *Stream {
	s := &Stream{name: name, sink: sink, hash: sha256.New()}
	if retain {
		s.keep = &bytes.Buffer{}
	}
	return s
}

// Append writes one record. Replayable records take the next seq and the given
// world time; a session record keeps the seq of the replayable record that
// follows it and the time of the one before.
func (s *Stream) Append(recordType string, atNS int64, fields map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.err != nil {
		return
	}
	rec := s.next(recordType, atNS, fields)
	line, err := encodeRecord(rec)
	if err != nil {
		s.err = err
		return
	}
	s.commit(rec, append(line, '\n'))
}

// next is the record the stream writes next: a session record keeps the seq of
// the replayable record that follows it and the time of the one before.
func (s *Stream) next(recordType string, atNS int64, fields map[string]any) Record {
	if IsSession(recordType) {
		atNS = s.atNS
	}
	return Record{Stream: s.name, Seq: s.seq, Type: recordType, AtNS: atNS, Fields: fields}
}

func (s *Stream) commit(rec Record, line []byte) {
	if !IsSession(rec.Type) {
		s.hash.Write(line)
		s.seq++
		s.atNS = rec.AtNS
	}
	if s.keep != nil {
		s.keep.Write(line)
	}
	if s.sink != nil {
		s.fail(s.sink.Write(line))
	}
}

// Flush pushes buffered lines to the destination and returns the first failure
// the stream has met.
func (s *Stream) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil && s.sink != nil && !s.closed {
		s.fail(s.sink.Flush())
	}
	return s.err
}

// Close ends the stream: it syncs and closes the destination and returns the
// stream's digest with the first failure. Closing twice returns the same.
func (s *Stream) Close() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed && s.sink != nil {
		s.fail(s.sink.Close())
	}
	s.closed = true
	return s.digest(), s.err
}

// Digest is the SHA-256 of the replayable lines written so far.
func (s *Stream) Digest() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.digest()
}

// Bytes returns the retained lines (nil unless the stream retains them).
func (s *Stream) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.keep == nil {
		return nil
	}
	return bytes.Clone(s.keep.Bytes())
}

func (s *Stream) digest() string {
	return "sha256:" + hex.EncodeToString(s.hash.Sum(nil))
}

func (s *Stream) fail(err error) {
	if err != nil && s.err == nil {
		s.err = fmt.Errorf("journal: %s stream: %w", s.name, err)
	}
}
