// Package uds is the plant wire node's transport edge: the host's owner-only unix socket and the capped wait for records.
package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var ErrSocketNotPrivate = errors.New("plantwire: the host socket is not owner-only")

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
	if err := ownerOnly(path); err != nil {
		return nil, err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("plantwire: dial %s: %w", path, err)
	}
	link := &Link{conn: conn, writer: bufio.NewWriter(conn), incoming: make(chan domain.Inbound, 64), done: make(chan struct{}), wait: wait}
	link.stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
	go link.read()
	return link, nil
}

func ownerOnly(path string) error {
	if err := owned(filepath.Dir(path), os.ModeDir); err != nil {
		return err
	}
	return owned(path, os.ModeSocket)
}

func owned(target string, kind os.FileMode) error {
	info, err := os.Lstat(target)
	if err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	owner, ok := info.Sys().(*syscall.Stat_t)
	if info.Mode().Perm()&0o077 != 0 || info.Mode().Type() != kind || !ok || int(owner.Uid) != os.Geteuid() {
		return fmt.Errorf("%w: %s (mode %v)", ErrSocketNotPrivate, target, info.Mode())
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
