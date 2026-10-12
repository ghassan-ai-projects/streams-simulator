package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAHostServesOneNodeOverAnOwnerOnlySocketAndWaitsBeforeEachSync(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	ready := make(chan struct{})
	served := make(chan error, 1)
	receive := func(map[string]any) []map[string]any { return []map[string]any{{"type": "HSYNC"}} }
	go func() { served <- ServeHost(t.Context(), path, receive, 20*time.Millisecond, func() { close(ready) }) }()
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
	err := ServeHost(ctx, testsupport.SocketPath(t), nil, 0, cancel)
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
	if err := ServeHost(t.Context(), filepath.Join(blocker, "h.sock"), nil, 0, nil); err == nil || !strings.Contains(err.Error(), "plantwire host:") {
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
	if err := ServeHost(t.Context(), filepath.Join(shared, "h.sock"), nil, 0, nil); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("ServeHost in a shared directory = %v", err)
	}
}
