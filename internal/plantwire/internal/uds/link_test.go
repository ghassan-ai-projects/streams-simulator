package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func privateListener(t *testing.T) (string, net.Listener) {
	t.Helper()
	path := testsupport.SocketPath(t)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	return path, listener
}

func echoHost(t *testing.T, listener net.Listener, reply map[string]any) {
	t.Helper()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			raw, _ := json.Marshal(reply)
			_, _ = conn.Write(append(raw, '\n'))
		}
	}()
}

func TestALinkSendsRecordsAndAwaitsTheHostsAnswer(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	echoHost(t, listener, map[string]any{"type": "HSYNC", "protocol_version": 1, "session_id": "s-1", "seq": 0, "lane": "control", "through_seq": 0})
	link, err := Dial(t.Context(), path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	if err := link.Send([]map[string]any{{"type": "NSYNC"}}); err != nil {
		t.Fatal(err)
	}
	records, met, err := link.Await(t.Context(), func(m domain.Message) bool { return m.Type == "HSYNC" })
	if err != nil || !met || len(records) != 1 {
		t.Fatalf("await = %v, %v, %v", records, met, err)
	}
}

func TestASilentHostLetsTheWaitExpire(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			<-release
			_ = conn.Close()
		}
	}()
	link, err := Dial(t.Context(), path, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	if records, met, err := link.Await(t.Context(), func(domain.Message) bool { return true }); met || err != nil || len(records) != 0 {
		t.Fatalf("await = %v, %v, %v", records, met, err)
	}
	if waited := link.WaitedMS(); waited < 50 {
		t.Fatalf("waited %d ms, want the whole cap", waited)
	}
}

func TestAClosedLinkCannotSend(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	echoHost(t, listener, map[string]any{"type": "HSYNC"})
	link, err := Dial(t.Context(), path, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := link.Close(); err != nil {
		t.Fatal(err)
	}
	if err := link.Send([]map[string]any{{"type": "NSYNC"}}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("send on a closed link = %v", err)
	}
}

func TestASocketOthersCanReachIsRefused(t *testing.T) {
	t.Parallel()
	path, _ := privateListener(t)
	if err := os.Chmod(path, 0o666); err != nil { //nolint:gosec // the test opens the socket to everyone on purpose
		t.Fatal(err)
	}
	if _, err := Dial(t.Context(), path, time.Second); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("dial = %v, want a refusal of a world-writable socket", err)
	}
	other, _ := privateListener(t)
	if err := os.Chmod(filepath.Dir(other), 0o755); err != nil { //nolint:gosec // the test shares the directory on purpose
		t.Fatal(err)
	}
	if _, err := Dial(t.Context(), other, time.Second); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("dial = %v, want a refusal of a socket in a shared directory", err)
	}
}

func TestAWaitEndsWhenTheHostHangsUpOrTheCallerCancels(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	go func() {
		if conn, err := listener.Accept(); err == nil {
			_ = conn.Close()
		}
	}()
	link, err := Dial(t.Context(), path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, _, err := link.Await(t.Context(), func(domain.Message) bool { return false }); !errors.Is(err, io.EOF) {
			t.Fatalf("await = %v, want the end of the host connection every time", err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := link.Await(ctx, func(domain.Message) bool { return false }); !errors.Is(err, context.Canceled) {
		t.Fatalf("await = %v, want the cancellation", err)
	}
	if err := link.Close(); err != nil {
		t.Fatal(err)
	}
	if err := link.Send([]map[string]any{{"type": "NSYNC"}}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("send = %v, want a closed link", err)
	}
	if err := link.Close(); err != nil {
		t.Fatalf("a second close = %v, want nothing to report", err)
	}
}

func TestOnlyAnOwnedSocketInAnOwnedDirectoryIsAccepted(t *testing.T) {
	t.Parallel()
	path, _ := privateListener(t)
	plain := filepath.Join(filepath.Dir(path), "plain")
	if err := os.WriteFile(plain, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Dial(t.Context(), plain, time.Second); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("dial a plain file = %v", err)
	}
	if _, err := Dial(t.Context(), filepath.Join(filepath.Dir(path), "missing.sock"), time.Second); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dial a missing socket = %v", err)
	}
}

func TestARecordThatCannotBeEncodedIsNotSent(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	echoHost(t, listener, map[string]any{"type": "HSYNC"})
	link, err := Dial(t.Context(), path, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	var unsupported *json.UnsupportedValueError
	if err := link.Send([]map[string]any{{"type": "DDATA", "value": math.NaN()}}); !errors.As(err, &unsupported) {
		t.Fatalf("send of NaN = %v", err)
	}
}
