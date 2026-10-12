package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAHostServesOneNodeOverAnOwnerOnlySocketAndWaitsBeforeEachSync(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	ready := make(chan struct{})
	served := make(chan error, 1)
	receive := &stubPeer{replies: []map[string]any{{"type": "HSYNC"}}}
	go func() {
		served <- ServeHost(t.Context(), path, receive, Pacing{Delay: 20 * time.Millisecond, Handshake: time.Second}, func() { close(ready) })
	}()
	<-ready
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket mode = %v, %v", info, err)
	}
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if _, err := conn.Write([]byte("not json\n{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	var reply map[string]any
	if err != nil || json.Unmarshal(line, &reply) != nil || reply["type"] != "HSYNC" {
		t.Fatalf("reply = %q, %v", line, err)
	}
	if waited := time.Since(started); waited < 20*time.Millisecond {
		t.Fatalf("waited %v before HSYNC", waited)
	}
	_ = conn.Close()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestAHostStopsWaitingForANodeWhenCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	err := ServeHost(ctx, testsupport.SocketPath(t), nil, Pacing{}, cancel)
	if err == nil || !errors.Is(err, net.ErrClosed) {
		t.Fatalf("ServeHost = %v, want a closed listener", err)
	}
}

func TestAHostCannotListenUnderAFile(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ServeHost(t.Context(), filepath.Join(blocker, "h.sock"), nil, Pacing{}, nil); err == nil || !strings.Contains(err.Error(), "plantwire host:") {
		t.Fatal("ServeHost under a file succeeded")
	}
}

func TestACancelledPauseStopsBeforeTheSync(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := pause(ctx, map[string]any{"type": "HSYNC"}, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("pause = %v", err)
	}
}

func TestAHostRefusesASharedDirectory(t *testing.T) {
	t.Parallel()
	shared := filepath.Join(t.TempDir(), "shared")
	if err := os.Mkdir(shared, 0o755); err != nil { //nolint:gosec // the test shares the directory on purpose
		t.Fatal(err)
	}
	if err := os.Chmod(shared, 0o755); err != nil { //nolint:gosec // the test shares the directory on purpose
		t.Fatal(err)
	}
	if err := ServeHost(t.Context(), filepath.Join(shared, "h.sock"), nil, Pacing{}, nil); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("ServeHost in a shared directory = %v", err)
	}
}

func TestAHostAnswerThatCannotBeEncodedEndsTheSession(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	ready := make(chan struct{})
	served := make(chan error, 1)
	receive := &stubPeer{replies: []map[string]any{{"type": "STATE", "heartbeat": math.Inf(1)}}}
	go func() {
		served <- ServeHost(t.Context(), path, receive, Pacing{Handshake: time.Second}, func() { close(ready) })
	}()
	<-ready
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write([]byte("{\"type\":\"NBIRTH\"}\n")); err != nil {
		t.Fatal(err)
	}
	var unsupported *json.UnsupportedValueError
	if err := <-served; !errors.As(err, &unsupported) {
		t.Fatalf("ServeHost = %v", err)
	}
}

type stubPeer struct {
	replies  []map[string]any
	accepted bool
	done     bool
}

func (p *stubPeer) Receive(domain.Inbound) []map[string]any { return p.replies }

func (p *stubPeer) Accepted() bool { return p.accepted }

func (p *stubPeer) Done() bool { return p.done }
