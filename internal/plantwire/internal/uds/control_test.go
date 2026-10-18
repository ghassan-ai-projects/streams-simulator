package uds

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type echoHandler struct {
	mu    sync.Mutex
	lines []string
}

func (h *echoHandler) Handle(_ context.Context, line []byte) []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lines = append(h.lines, string(line))
	return []byte(`{"echo":"` + string(line) + `"}`)
}

func (h *echoHandler) seen() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.lines...)
}

func listenControl(t *testing.T) (string, *ControlServer, *echoHandler) {
	t.Helper()
	path, handler := testsupport.SocketPath(t), &echoHandler{}
	server, err := ListenControl(t.Context(), path, handler)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return path, server, handler
}

func TestTheControlSocketIsOwnerOnlyAnswersEachRequestLineWithOneLineAndIsRemovedOnClose(t *testing.T) {
	t.Parallel()
	path, server, handler := listenControl(t)
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket = %v, %v, want mode 0600", info, err)
	}
	conn := dialHost(t, path)
	reader := bufio.NewReader(conn)
	for _, request := range []string{"one", "two"} {
		if _, err := conn.Write([]byte(request + "\n")); err != nil {
			t.Fatal(err)
		}
		line, err := reader.ReadString('\n')
		if err != nil || line != `{"echo":"`+request+`"}`+"\n" {
			t.Fatalf("reply to %q = %q, %v, want one line", request, line, err)
		}
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat after Close = %v, want the socket removed", err)
	}
	if seen := handler.seen(); len(seen) != 2 || seen[0] != "one" || seen[1] != "two" {
		t.Fatalf("handled %v, want both requests in order", seen)
	}
}

func requireClosed(t *testing.T, err error, what string) {
	t.Helper()
	if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("%s: read = %v, want the connection closed by the server", what, err)
	}
}

func TestTheControlSocketRefusesAPathThatIsInUse(t *testing.T) {
	t.Parallel()
	path, _, handler := listenControl(t)
	if _, err := ListenControl(t.Context(), path, handler); !errors.Is(err, ErrSocketInUse) {
		t.Fatalf("ListenControl on a live socket = %v, want ErrSocketInUse", err)
	}
}

func TestTheControlSocketTakesNoMoreConnectionsThanItsBoundAndAnOverlongLineEndsOnlyItsOwn(t *testing.T) {
	t.Parallel()
	path, _, _ := listenControl(t)
	held := make([]net.Conn, 0, maxControlConns)
	for range maxControlConns {
		conn := dialHost(t, path)
		if _, err := conn.Write([]byte("hold\n")); err != nil {
			t.Fatal(err)
		}
		if _, err := bufio.NewReader(conn).ReadString('\n'); err != nil {
			t.Fatal(err)
		}
		held = append(held, conn)
	}
	over := dialHost(t, path)
	_, err := bufio.NewReader(over).ReadString('\n')
	requireClosed(t, err, "a connection over the bound")
	long := strings.Repeat("x", maxControlLine+1) + "\n"
	if _, err := held[0].Write([]byte(long)); err != nil {
		t.Fatal(err)
	}
	_, err = bufio.NewReader(held[0]).ReadString('\n')
	requireClosed(t, err, "an overlong request")
	if _, err := held[1].Write([]byte("still\n")); err != nil {
		t.Fatal(err)
	}
	if line, err := bufio.NewReader(held[1]).ReadString('\n'); err != nil || line != `{"echo":"still"}`+"\n" {
		t.Fatalf("another connection = %q, %v, want it unaffected", line, err)
	}
}
