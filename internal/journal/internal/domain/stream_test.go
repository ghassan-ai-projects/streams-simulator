package domain

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

// memorySink is a Sink that keeps its lines and can be told to fail.
type memorySink struct {
	lines              []string
	writeErr, flushErr error
	closeErr           error
	flushes            int
	closed             bool
}

func (m *memorySink) Write(line []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.lines = append(m.lines, string(line))
	return nil
}

func (m *memorySink) Flush() error { m.flushes++; return m.flushErr }

func (m *memorySink) Close() error { m.closed = true; return m.closeErr }

func TestStreamNumbersReplayableRecordsWithoutGaps(t *testing.T) {
	t.Parallel()
	sink := &memorySink{}
	s := NewStream(StreamObserver, sink, false)
	s.Append(TypeClockAdvanced, 10, map[string]any{"to_ns": 10})
	s.Append(TypeConsumerReported, 0, map[string]any{"quiesced_through_ns": 10, "verdict_submitted": false})
	s.Append(TypeClockAdvanced, 20, map[string]any{"to_ns": 20})
	records, err := Read([]byte(strings.Join(sink.lines, "")))
	if err != nil || len(records) != 3 {
		t.Fatalf("read = %d records, %v", len(records), err)
	}
	// The session record keeps the seq of the replayable record that follows
	// it and the time of the one before.
	got := [][2]int64{{records[0].Seq, records[0].AtNS}, {records[1].Seq, records[1].AtNS}, {records[2].Seq, records[2].AtNS}}
	want := [][2]int64{{0, 10}, {1, 10}, {1, 20}}
	if !slices.Equal(got, want) {
		t.Fatalf("seq/at_ns = %v, want %v", got, want)
	}
}

func TestStreamDigestIgnoresSessionRecords(t *testing.T) {
	t.Parallel()
	with, without := NewStream(StreamObserver, nil, true), NewStream(StreamObserver, nil, true)
	for _, s := range []*Stream{with, without} {
		s.Append(TypeClockAdvanced, 10, map[string]any{"to_ns": 10})
	}
	with.Append(TypeConsumerReported, 0, map[string]any{"quiesced_through_ns": 10, "verdict_submitted": true})
	for _, s := range []*Stream{with, without} {
		s.Append(TypeClockAdvanced, 20, map[string]any{"to_ns": 20})
	}
	if with.Digest() != without.Digest() {
		t.Fatalf("a session record changed the digest: %s vs %s", with.Digest(), without.Digest())
	}
	if Digest(with.Bytes()) != with.Digest() || Digest(without.Bytes()) != without.Digest() {
		t.Fatal("the digest of the bytes must equal the digest the stream reports")
	}
	if len(with.Bytes()) <= len(without.Bytes()) {
		t.Fatal("the session record must still be written to the stream")
	}
}

func TestStreamRemembersAnEncodeFailureAndStopsWriting(t *testing.T) {
	t.Parallel()
	sink := &memorySink{}
	s := NewStream(StreamDirector, sink, false)
	s.Append(TypeCommand, 1, map[string]any{"command": map[string]any{"args": math.NaN()}})
	s.Append(TypeCommand, 2, map[string]any{"command": "after"})
	err := s.Flush()
	if err == nil || !strings.Contains(err.Error(), "encode command record") {
		t.Fatalf("Flush = %v, want the encode failure", err)
	}
	if _, closeErr := s.Close(); closeErr == nil || len(sink.lines) != 0 {
		t.Fatalf("Close = %v with %d lines written, want the failure and no lines", closeErr, len(sink.lines))
	}
}

func TestStreamReturnsSinkFailures(t *testing.T) {
	t.Parallel()
	boom := errors.New("disk full")
	cases := map[string]*memorySink{
		"write": {writeErr: boom}, "flush": {flushErr: boom}, "close": {closeErr: boom},
	}
	for name, sink := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := NewStream(StreamObserver, sink, false)
			s.Append(TypeClockAdvanced, 1, map[string]any{"to_ns": 1})
			_ = s.Flush()
			if _, err := s.Close(); !errors.Is(err, boom) {
				t.Fatalf("Close = %v, want %v", err, boom)
			}
		})
	}
}

func TestStreamIgnoresRecordsAfterClose(t *testing.T) {
	t.Parallel()
	sink := &memorySink{}
	s := NewStream(StreamObserver, sink, false)
	s.Append(TypeClockAdvanced, 1, map[string]any{"to_ns": 1})
	digest, err := s.Close()
	if err != nil || !sink.closed {
		t.Fatalf("Close = %v, closed %v", err, sink.closed)
	}
	s.Append(TypeClockAdvanced, 2, map[string]any{"to_ns": 2})
	again, err := s.Close()
	if err != nil || again != digest || len(sink.lines) != 1 || s.Flush() != nil {
		t.Fatalf("after close: digest %s/%s, %d lines, err %v", again, digest, len(sink.lines), err)
	}
}

func TestStreamFlushReachesTheSink(t *testing.T) {
	t.Parallel()
	sink := &memorySink{}
	s := NewStream(StreamObserver, sink, false)
	if err := s.Flush(); err != nil || sink.flushes != 1 {
		t.Fatalf("Flush = %v after %d sink flushes", err, sink.flushes)
	}
	if s.Bytes() != nil {
		t.Fatal("a stream that does not retain returns no bytes")
	}
}

var errFlush = errors.New("flush refused")
