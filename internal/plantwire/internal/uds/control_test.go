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
	"time"

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

func openControl(t *testing.T) (string, *ControlServer, *echoHandler) {
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
	path, server, handler := openControl(t)
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
	path, _, handler := openControl(t)
	if _, err := ListenControl(t.Context(), path, handler); !errors.Is(err, ErrSocketInUse) {
		t.Fatalf("ListenControl on a live socket = %v, want ErrSocketInUse", err)
	}
}

func TestTheControlSocketTakesNoMoreConnectionsThanItsBoundAndAnOverlongLineEndsOnlyItsOwn(t *testing.T) {
	t.Parallel()
	path, _, _ := openControl(t)
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

func TestTheControlSocketTakesALineOfExactlyTheBridgesLongestRequest(t *testing.T) {
	t.Parallel()
	path, _, _ := openControl(t)
	conn := dialHost(t, path)
	line := `{"op":"state"}` + strings.Repeat(" ", maxControlLine-len(`{"op":"state"}`))
	if _, err := conn.Write([]byte(line + "\n")); err != nil {
		t.Fatal(err)
	}
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || reply != `{"echo":"`+line+`"}`+"\n" {
		t.Fatalf("reply to a line of %d bytes = %q, %v, want it handled: the bridge sends up to that", len(line), reply, err)
	}
}

func TestAControlConnectionThatSaysNothingIsClosedAfterItsIdleTime(t *testing.T) {
	t.Parallel()
	path, handler := testsupport.SocketPath(t), &echoHandler{}
	server, err := listenControl(t.Context(), path, handler, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	conn := dialHost(t, path)
	_, err = bufio.NewReader(conn).ReadString('\n')
	requireClosed(t, err, "an idle connection")
	if len(handler.seen()) != 0 {
		t.Fatalf("handled %v, want nothing: the connection never spoke", handler.seen())
	}
}
