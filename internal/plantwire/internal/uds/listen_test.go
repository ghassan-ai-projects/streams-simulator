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

func TestAHostKeepsHeartbeatingWhileItHoldsItsAnswer(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	ready := make(chan struct{})
	peer := &stubPeer{replies: []map[string]any{{"type": "HSYNC"}}, beat: map[string]any{"type": "STATE"}}
	go func() {
		_ = ServeHost(t.Context(), path, peer, Pacing{Delay: 300 * time.Millisecond, Handshake: time.Second, Heartbeat: 50 * time.Millisecond}, func() { close(ready) })
	}()
	<-ready
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	reader, beats := bufio.NewReader(conn), 0
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatalf("the answer never came: %v", err)
		}
		var record map[string]any
		_ = json.Unmarshal(line, &record)
		if record["type"] == "HSYNC" {
			break
		}
		beats++
	}
	if beats < 3 {
		t.Fatalf("%d heartbeats while the answer was held, want at least 3", beats)
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

func TestACancelledHoldStopsBeforeTheSync(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if open, err := hold(ctx, nil, &stubPeer{}, nil, time.Hour); open || !errors.Is(err, context.Canceled) {
		t.Fatalf("hold = %v, %v", open, err)
	}
}

func TestAHostRefusesASharedDirectory(t *testing.T) {
	t.Parallel()
	shared := filepath.Join(t.TempDir(), "shared")
	if err := os.Mkdir(shared, 0o700); err != nil {
		t.Fatal(err)
	}
	widen(t, filepath.Dir(shared), "shared", 0o755)
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
	replies     []map[string]any
	beat        map[string]any
	beats       int
	accepted    bool
	done        bool
	doneOnFirst bool
}

func (p *stubPeer) Receive(domain.Inbound) []map[string]any {
	p.done = p.done || p.doneOnFirst
	return p.replies
}

func (p *stubPeer) Beat() []map[string]any {
	if p.beat == nil {
		return nil
	}
	p.beats++
	return []map[string]any{{"type": p.beat["type"], "heartbeat": p.beats}}
}

func (p *stubPeer) Accepted() bool { return p.accepted }

func (p *stubPeer) Done() bool { return p.done }

func TestAHostWhoseNodeHangsUpDuringABarrierWaitEndsWithThePeerGone(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	served := serveInBackground(t, path, &stubPeer{replies: []map[string]any{{"type": "HSYNC"}}}, Pacing{Delay: 100 * time.Millisecond, Handshake: time.Second})
	conn := dialHost(t, path)
	if _, err := conn.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	if err := <-served; !PeerGone(err) || Interrupted(err) {
		t.Fatalf("ServeHost = %v, want the node gone, not an interrupt", err)
	}
}

func TestAHostInterruptedBeforeANodeEndsInterruptedNotWithThePeerGone(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	if err := ServeHost(ctx, testsupport.SocketPath(t), nil, Pacing{}, cancel); !Interrupted(err) || PeerGone(err) {
		t.Fatalf("ServeHost = %v, want an interrupt", err)
	}
}
