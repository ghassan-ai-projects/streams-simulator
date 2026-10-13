package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
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

var ErrSocketInUse = kernel.ErrInUse

func ServeHost(ctx context.Context, path string, peer Peer, pacing Pacing, ready func()) error {
	ctx, stop := kernel.UntilInterrupted(ctx)
	defer stop()
	listener, err := kernel.ListenPrivate(ctx, path)
	if err != nil {
		return fmt.Errorf("plantwire host: %w", err)
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
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("plantwire host: accept: %w", err)
		}
		if spoke, err := session(ctx, conn, peer, pacing); spoke || err != nil || peer.Done() {
			return err
		}
	}
}

func session(ctx context.Context, conn net.Conn, peer Peer, pacing Pacing) (bool, error) {
	defer func() { _ = conn.Close() }()
	hangUp := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer hangUp()
	if err := conn.SetReadDeadline(time.Now().Add(pacing.Handshake)); err != nil {
		return true, fmt.Errorf("plantwire host: %w", err)
	}
	return serveConn(ctx, conn, peer, pacing.Delay)
}

func serveConn(ctx context.Context, conn net.Conn, peer Peer, delay time.Duration) (bool, error) {
	writer, scanner := bufio.NewWriter(conn), lines(conn)
	spoke := false
	for !peer.Done() && scanner.Scan() {
		spoke = true
		if err := respond(ctx, conn, writer, peer, scanner.Bytes(), delay); err != nil {
			return true, err
		}
	}
	if err := scanner.Err(); err != nil {
		return true, fmt.Errorf("plantwire host: read: %w", err)
	}
	return spoke, nil
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
