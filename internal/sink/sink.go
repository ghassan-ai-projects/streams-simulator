// Package sink delivers rendered records. Three sinks: inproc (a buffer),
// file (byte-reproducible), and http-push (stepped: deterministic; wall:
// real-time withholding, excluded from the deterministic build via the
// simdet tag). Evidence leaves on a sink; MCP never carries it.
package sink

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

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

// File writes lines to a path, buffered; Close flushes.
type File struct {
	path string
	f    *os.File
	w    *bufio.Writer
	mu   sync.Mutex
}

// NewFile opens the sink at path (created if missing).
func NewFile(path string) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil && filepath.Dir(path) != "." {
		if err != nil {
			return nil, fmt.Errorf("sink: mkdir: %w", err)
		}
		return nil, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("sink: create %s: %w", path, err)
	}
	return &File{path: path, f: f, w: bufio.NewWriter(f)}, nil
}

// Write appends a line.
func (s *File) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.w.Write(line); err != nil {
		return fmt.Errorf("sink: write %s: %w", s.path, err)
	}
	if err := s.w.WriteByte('\n'); err != nil {
		return fmt.Errorf("sink: write %s: %w", s.path, err)
	}
	return nil
}

// Close flushes and returns the file contents.
func (s *File) Close() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.w.Flush(); err != nil {
		return nil, fmt.Errorf("sink: flush %s: %w", s.path, err)
	}
	if err := s.f.Sync(); err != nil {
		return nil, fmt.Errorf("sink: sync %s: %w", s.path, err)
	}
	if err := s.f.Close(); err != nil {
		return nil, fmt.Errorf("sink: close %s: %w", s.path, err)
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("sink: read back %s: %w", s.path, err)
	}
	return raw, nil
}

// HTTPPush posts each line to an endpoint. In stepped mode delivery is
// immediate and in order, so the sink is byte-deterministic. The wall
// sub-mode (real-time withholding by observed_time) lives behind the simdet
// build tag and is unavailable to the deterministic test suite.
type HTTPPush struct {
	url    string
	client *http.Client
	mu     sync.Mutex
	buf    bytes.Buffer
	ctx    context.Context
}

// NewHTTPPush builds the sink. ctx cancels the run.
func NewHTTPPush(ctx context.Context, url string) *HTTPPush {
	return &HTTPPush{
		url:    url,
		client: &http.Client{Timeout: 30 * time.Second},
		ctx:    ctx,
	}
}

// Write posts one line, recording it for the digest.
func (s *HTTPPush) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.url, bytes.NewReader(line))
	if err != nil {
		return fmt.Errorf("sink: http-push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/jsonl")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sink: http-push POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sink: http-push POST %s: status %d", s.url, resp.StatusCode)
	}
	s.buf.Write(line)
	s.buf.WriteByte('\n')
	return nil
}

// Close returns the delivered bytes (in delivery order).
func (s *HTTPPush) Close() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, s.buf.Len())
	copy(out, s.buf.Bytes())
	return out, nil
}
