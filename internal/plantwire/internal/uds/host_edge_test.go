package uds

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func serveInBackground(t *testing.T, path string, peer Peer, pacing Pacing) chan error {
	t.Helper()
	ready, served := make(chan struct{}), make(chan error, 1)
	go func() { served <- ServeHost(t.Context(), path, peer, pacing, func() { close(ready) }) }()
	select {
	case <-ready:
	case err := <-served:
		t.Fatalf("the host did not start: %v", err)
	}
	return served
}

func dialHost(t *testing.T, path string) net.Conn {
	t.Helper()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestAHostReplacesAStaleSocketButNotALiveHostOrAFile(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	stale, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	stale.(*net.UnixListener).SetUnlinkOnClose(false)
	_ = stale.Close()
	served := serveInBackground(t, path, &stubPeer{doneOnFirst: true}, Pacing{Handshake: time.Second})
	if err := ServeHost(t.Context(), path, &stubPeer{}, Pacing{}, nil); !errors.Is(err, ErrSocketInUse) {
		t.Fatalf("a second host on a live socket = %v", err)
	}
	if info, err := os.Lstat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("the published socket = %v, %v", info, err)
	}
	_, _ = dialHost(t, path).Write([]byte("{\"type\":\"NBIRTH\"}\n"))
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(filepath.Dir(path), "plain")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ServeHost(t.Context(), file, &stubPeer{}, Pacing{}, nil); !errors.Is(err, ErrSocketInUse) {
		t.Fatalf("a host on a regular file = %v", err)
	}
}

func TestAHostThatIsDoneClosesTheConnection(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	served := serveInBackground(t, path, &stubPeer{doneOnFirst: true}, Pacing{Handshake: time.Second})
	conn := dialHost(t, path)
	if _, err := conn.Write([]byte("{\"type\":\"NBIRTH\"}\n{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	if err := <-served; err != nil {
		t.Fatalf("a refusing host = %v", err)
	}
	if _, err := bufio.NewReader(conn).ReadBytes('\n'); !errors.Is(err, io.EOF) {
		t.Fatalf("read on the refused connection = %v, want its end", err)
	}
}

func TestAnIdlePeerCannotHoldTheHostPastTheHandshake(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	served := serveInBackground(t, path, &stubPeer{}, Pacing{Handshake: 50 * time.Millisecond})
	dialHost(t, path)
	var timeout net.Error
	if err := <-served; !errors.As(err, &timeout) || !timeout.Timeout() {
		t.Fatalf("an idle peer = %v, want a handshake timeout", err)
	}
}

func TestAnAcceptedPeerHasNoReadDeadline(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	peer := &stubPeer{accepted: true, replies: []map[string]any{{"type": "STATE"}}}
	served := serveInBackground(t, path, peer, Pacing{Handshake: 30 * time.Millisecond})
	conn := dialHost(t, path)
	reader := bufio.NewReader(conn)
	for range 2 {
		if _, err := conn.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
			t.Fatal(err)
		}
		if _, err := reader.ReadBytes('\n'); err != nil {
			t.Fatal(err)
		}
		<-time.After(60 * time.Millisecond)
	}
	_ = conn.Close()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestAFloodingHostIsCutOffAtOneBarrier(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		line, _ := json.Marshal(map[string]any{"type": "DQUERY"})
		for range MaxRecordsPerWait + 1 {
			if _, err := conn.Write(append(line, '\n')); err != nil {
				return
			}
		}
	}()
	link, err := Dial(t.Context(), path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	records, met, err := link.Await(t.Context(), func(domain.Message) bool { return false })
	if !errors.Is(err, ErrFlooded) || met || len(records) != MaxRecordsPerWait {
		t.Fatalf("await = %d records, %v, %v", len(records), met, err)
	}
}
