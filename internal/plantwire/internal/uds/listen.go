package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type Peer interface {
	Receive(domain.Inbound) []map[string]any
	Accepted() bool
	Done() bool
}

type Pacing struct {
	Delay     time.Duration
	Handshake time.Duration
}

const HandshakeWait = 10 * time.Second

var ErrSocketInUse = errors.New("plantwire host: the socket path is in use")

func ServeHost(ctx context.Context, path string, peer Peer, pacing Pacing, ready func()) error {
	ctx, stop := UntilInterrupted(ctx)
	defer stop()
	listener, err := listenPrivately(ctx, path)
	if err != nil {
		return err
	}
	defer func() { _, _ = listener.Close(), os.Remove(path) }()
	if ready != nil {
		ready()
	}
	return serveOne(ctx, listener, peer, pacing)
}

func serveOne(ctx context.Context, listener net.Listener, peer Peer, pacing Pacing) error {
	unblock := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer unblock()
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("plantwire host: accept: %w", err)
	}
	defer func() { _ = conn.Close() }()
	hangUp := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer hangUp()
	if err := conn.SetReadDeadline(time.Now().Add(pacing.Handshake)); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	return serveConn(ctx, conn, peer, pacing.Delay)
}

func serveConn(ctx context.Context, conn net.Conn, peer Peer, delay time.Duration) error {
	writer, scanner := bufio.NewWriter(conn), lines(conn)
	for !peer.Done() && scanner.Scan() {
		if err := respond(ctx, conn, writer, peer, scanner.Bytes(), delay); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("plantwire host: read: %w", err)
	}
	return nil
}

func lines(conn net.Conn) *bufio.Scanner {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), domain.MaxBirthBytes)
	return scanner
}

func respond(ctx context.Context, conn net.Conn, writer *bufio.Writer, peer Peer, line []byte, delay time.Duration) error {
	in, ok := domain.Decode(line)
	if !ok {
		return nil
	}
	if err := answer(ctx, writer, peer.Receive(in), delay); err != nil {
		return err
	}
	return settle(conn, peer)
}

func settle(conn net.Conn, peer Peer) error {
	if !peer.Accepted() {
		return nil
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	return nil
}

func answer(ctx context.Context, writer *bufio.Writer, records []map[string]any, delay time.Duration) error {
	for _, record := range records {
		if err := pause(ctx, record, delay); err != nil {
			return err
		}
		if err := writeLine(writer, record); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("plantwire host: write: %w", err)
	}
	return nil
}

func pause(ctx context.Context, record map[string]any, delay time.Duration) error {
	if record["type"] != "HSYNC" || delay <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("plantwire host: %w", ctx.Err())
	case <-time.After(delay):
		return nil
	}
}

func writeLine(writer *bufio.Writer, record map[string]any) error {
	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("plantwire host: encode: %w", err)
	}
	if _, err := writer.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("plantwire host: write: %w", err)
	}
	return nil
}

func listenPrivately(ctx context.Context, path string) (net.Listener, error) {
	if err := prepare(ctx, path); err != nil {
		return nil, err
	}
	binding := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".binding")
	listener, err := bindAt(ctx, binding)
	if err != nil {
		return nil, err
	}
	if err := publish(binding, path); err != nil {
		_ = listener.Close()
		return nil, err
	}
	return listener, nil
}

func prepare(ctx context.Context, path string) error {
	if err := privateDir(filepath.Dir(path)); err != nil {
		return err
	}
	return clearStale(ctx, path)
}

func privateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	if err := owned(dir, os.ModeDir); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	return nil
}

func clearStale(ctx context.Context, path string) error {
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("plantwire host: %w", err)
	case info.Mode().Type() != os.ModeSocket:
		return fmt.Errorf("%w: %s is not a socket", ErrSocketInUse, path)
	case listening(ctx, path):
		return fmt.Errorf("%w: a host already listens on %s", ErrSocketInUse, path)
	}
	return removed(path)
}

func listening(ctx context.Context, path string) bool {
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func removed(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	return nil
}

func bindAt(ctx context.Context, binding string) (*net.UnixListener, error) {
	_ = os.Remove(binding)
	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", binding)
	if err != nil {
		return nil, fmt.Errorf("plantwire host: listen %s: %w", binding, err)
	}
	unix := listener.(*net.UnixListener)
	unix.SetUnlinkOnClose(false)
	return unix, nil
}

func publish(binding, path string) error {
	if err := os.Chmod(binding, 0o600); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	if err := os.Rename(binding, path); err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	return nil
}
