// Package httppush is the network edge of the sink module: each line is
// posted to an endpoint in order, and the delivered bytes are kept for the
// digest.
package httppush

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HTTPPush posts each line to an endpoint. Delivery is immediate and in
// order, so the sink is byte-deterministic. There is no real-time (wall)
// sub-mode: withholding by observed time is the perturbation layer's job.
type HTTPPush struct {
	url    string
	client *http.Client
	mu     sync.Mutex
	buf    bytes.Buffer
	ctx    context.Context
}

// New builds the sink. ctx cancels the run.
func New(ctx context.Context, url string) *HTTPPush {
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
	resp, err := s.postLine(line)
	if err != nil {
		return err
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

func (s *HTTPPush) postLine(line []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.url, bytes.NewReader(line))
	if err != nil {
		return nil, fmt.Errorf("sink: http-push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/jsonl")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sink: http-push POST: %w", err)
	}
	return resp, nil
}
