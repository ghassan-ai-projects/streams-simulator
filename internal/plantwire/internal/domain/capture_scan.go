package domain

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

const (
	MaxScanBytes   = 32 << 20
	MaxLineBytes   = MaxBirthBytes + 4096
	maxLaterBirths = 64
	readChunk      = 64 << 10
)

type Sighting struct {
	Line   int
	Birth  []byte
	BDSeq  uint64
	Digest string
}

type Unreadable struct {
	Line  int
	Cause error
}

type Scan struct {
	Taken      *Sighting
	Later      []Sighting
	Seen       int
	Unreadable []Unreadable
	Truncated  bool
}

type scanner struct {
	lines  *lineReader
	head   *headBuffer
	number int
	scan   Scan
}

func ScanBirths(reader io.Reader) (Scan, error) {
	head := &headBuffer{}
	s := &scanner{lines: newLineReader(io.TeeReader(reader, head)), head: head}
	if err := s.run(); err != nil {
		return Scan{}, err
	}
	return s.finish()
}

func (s *scanner) run() error {
	for !s.scan.Truncated {
		done, err := s.step()
		if done || err != nil {
			return err
		}
	}
	return nil
}

func (s *scanner) step() (bool, error) {
	line, tooLong, err := s.lines.next()
	if err == nil || len(line) > 0 || tooLong {
		s.number++
		if refusal := s.take(line, tooLong); refusal != nil {
			return true, refusal
		}
	}
	s.scan.Truncated = err == nil && s.lines.read > MaxScanBytes
	return err != nil, readFailure(err)
}

func readFailure(err error) error {
	if err == nil || errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (s *scanner) take(line []byte, tooLong bool) error {
	birth, found, err := classify(line, tooLong)
	switch {
	case isStreamRefusal(err):
		return err
	case err != nil:
		s.skip(err)
	case found:
		s.sight(birth)
	}
	return nil
}

func (s *scanner) skip(cause error) {
	s.scan.Unreadable = append(s.scan.Unreadable, Unreadable{Line: s.number, Cause: cause})
}

func (s *scanner) sight(birth []byte) {
	s.scan.Seen++
	bdSeq, digest, err := contentOf(birth)
	sighting := Sighting{Line: s.number, BDSeq: bdSeq, Digest: digest}
	switch {
	case s.scan.Taken == nil:
		sighting.Birth = birth
		s.scan.Taken = &sighting
	case err != nil:
		s.skip(err)
	case len(s.scan.Later) < maxLaterBirths:
		s.scan.Later = append(s.scan.Later, sighting)
	}
}

func (s *scanner) finish() (Scan, error) {
	if s.scan.Seen == 0 && s.lines.read <= MaxBirthBytes {
		if err := s.wholeDocument(); err != nil {
			return Scan{}, err
		}
	}
	if s.scan.Taken == nil {
		return Scan{}, noBirth(s.scan)
	}
	return s.scan, nil
}

func (s *scanner) wholeDocument() error {
	s.number = 1
	birth, found, err := classify(s.head.data, false)
	switch {
	case isStreamRefusal(err):
		return err
	case found:
		s.scan.Unreadable = nil
		s.sight(birth)
	}
	return nil
}

func contentOf(birth []byte) (uint64, string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(birth, &fields); err != nil {
		return 0, "", fmt.Errorf("plantwire: not a JSON record: %w", err)
	}
	bdSeq, err := strconv.ParseUint(string(fields["bd_seq"]), 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("plantwire: bd_seq of the birth: %w", err)
	}
	digest, err := kindsDigest(fields["kinds"])
	return bdSeq, digest, err
}

func kindsDigest(raw json.RawMessage) (string, error) {
	var kinds any
	if err := decode(raw, &kinds); err != nil {
		return "", fmt.Errorf("plantwire: kinds of the birth: %w", err)
	}
	digest, err := canonical.Digest(kinds)
	if err != nil {
		return "", fmt.Errorf("plantwire: kinds of the birth: %w: %w", ErrNotCanonical, err)
	}
	return digest, nil
}

type headBuffer struct{ data []byte }

func (h *headBuffer) Write(p []byte) (int, error) {
	if room := MaxBirthBytes + 1 - len(h.data); room > 0 {
		h.data = append(h.data, p[:min(room, len(p))]...)
	}
	return len(p), nil
}

type lineReader struct {
	in   *bufio.Reader
	read int64
}

func newLineReader(source io.Reader) *lineReader {
	return &lineReader{in: bufio.NewReaderSize(source, readChunk)}
}

func (l *lineReader) next() ([]byte, bool, error) {
	var line []byte
	tooLong := false
	for {
		chunk, err := l.in.ReadSlice('\n')
		l.read += int64(len(chunk))
		if tooLong = tooLong || len(line)+len(chunk) > MaxLineBytes; !tooLong {
			line = append(line, chunk...)
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return finishedLine(line, tooLong), tooLong, readError(err)
		}
	}
}

func finishedLine(line []byte, tooLong bool) []byte {
	if tooLong {
		return nil
	}
	return bytes.TrimRight(line, "\r\n")
}

func readError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("plantwire: read: %w", err)
}
