package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sync"
)

type Sink interface {
	Write(line []byte) error
	Flush() error

	Close() error
}

type Stream struct {
	mu     sync.Mutex
	name   string
	sink   Sink
	hash   hash.Hash
	file   hash.Hash
	keep   *bytes.Buffer
	seq    int64
	atNS   int64
	err    error
	closed bool
}

func NewStream(name string, sink Sink, retain bool) *Stream {
	s := &Stream{name: name, sink: sink, hash: sha256.New(), file: sha256.New()}
	if retain {
		s.keep = &bytes.Buffer{}
	}
	return s
}

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
	s.file.Write(line)
	if s.keep != nil {
		s.keep.Write(line)
	}
	if s.sink != nil {
		s.fail(s.sink.Write(line))
	}
}

func (s *Stream) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil && s.sink != nil && !s.closed {
		s.fail(s.sink.Flush())
	}
	return s.err
}

func (s *Stream) Close() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed && s.sink != nil {
		s.fail(s.sink.Close())
	}
	s.closed = true
	return s.digest(), s.err
}

func (s *Stream) Digest() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.digest()
}

func (s *Stream) FileDigest() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return "sha256:" + hex.EncodeToString(s.file.Sum(nil))
}

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
