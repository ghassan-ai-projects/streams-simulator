// Package uds is the plant wire node's transport edge: the host's owner-only unix socket and the capped wait for records.
package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var ErrSocketNotPrivate = kernel.ErrNotPrivate

type Link struct {
	conn     net.Conn
	writer   *bufio.Writer
	incoming chan domain.Inbound
	done     chan struct{}
	wait     time.Duration
	stop     func() bool
	waited   int64
	failed   error
}

func Dial(ctx context.Context, path string, wait time.Duration) (*Link, error) {
	conn, err := kernel.DialPrivate(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("plantwire: %w", err)
	}
	link := NewAccepted(conn, wait)
	link.stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
	return link, nil
}

func NewAccepted(conn net.Conn, wait time.Duration) *Link {
	link := &Link{conn: conn, writer: bufio.NewWriter(conn), incoming: make(chan domain.Inbound, 64), done: make(chan struct{}), wait: wait}
	link.stop = func() bool { return false }
	go link.read()
	return link
}

func (l *Link) Incoming() <-chan domain.Inbound { return l.incoming }

func (l *Link) Expect(within time.Duration) error {
	if err := l.conn.SetReadDeadline(time.Now().Add(within)); err != nil {
		return fmt.Errorf("plantwire: read deadline: %w", err)
	}
	return nil
}

func (l *Link) Settle() error {
	if err := l.conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("plantwire: read deadline: %w", err)
	}
	return nil
}

func (l *Link) Close() error {
	l.stop()
	select {
	case <-l.done:
	default:
		close(l.done)
	}
	if err := l.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("plantwire: close: %w", err)
	}
	return nil
}

func (l *Link) Send(records []map[string]any) error {
	if err := l.conn.SetWriteDeadline(time.Now().Add(l.wait)); err != nil {
		return fmt.Errorf("plantwire: write: %w", err)
	}
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("plantwire: encode %v: %w", record["type"], err)
		}
		if _, err := l.writer.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("plantwire: write: %w", err)
		}
	}
	return l.flush()
}

func (l *Link) flush() error {
	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("plantwire: write: %w", err)
	}
	return nil
}
