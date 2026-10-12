package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type Receive func(record map[string]any) []map[string]any

func ServeHost(ctx context.Context, path string, receive Receive, delay time.Duration, ready func()) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := listenPrivately(ctx, path, ready)
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	return serveOne(ctx, listener, receive, delay)
}

func serveOne(ctx context.Context, listener net.Listener, receive Receive, delay time.Duration) error {
	unblock := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer unblock()
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("plantwire host: accept: %w", err)
	}
	defer func() { _ = conn.Close() }()
	hangUp := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer hangUp()
	return serveConn(ctx, conn, receive, delay)
}

func listenPrivately(ctx context.Context, path string, ready func()) (net.Listener, error) {
	listener, err := bind(ctx, path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("plantwire host: %w", err)
	}
	if ready != nil {
		ready()
	}
	return listener, nil
}

func serveConn(ctx context.Context, conn net.Conn, receive Receive, delay time.Duration) error {
	writer, scanner := bufio.NewWriter(conn), lines(conn)
	for scanner.Scan() {
		if record, _ := domain.Parse(scanner.Bytes()); record != nil {
			if err := answer(ctx, writer, receive(record), delay); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("plantwire host: read: %w", err)
	}
	return nil
}

func lines(conn net.Conn) *bufio.Scanner {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	return scanner
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

func bind(ctx context.Context, path string) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("plantwire host: %w", err)
	}
	if err := owned(filepath.Dir(path), os.ModeDir); err != nil {
		return nil, fmt.Errorf("plantwire host: %w", err)
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("plantwire host: listen %s: %w", path, err)
	}
	return listener, nil
}
