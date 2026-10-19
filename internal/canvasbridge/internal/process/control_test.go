package process

import (
	"bufio"
	"context"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func listenForControl(t *testing.T, answer func(request []byte, conn net.Conn)) string {
	t.Helper()
	path := testsupport.SocketPath(t)
	listener, err := kernel.ListenPrivate(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		request, _ := bufio.NewReader(conn).ReadBytes('\n')
		answer(request, conn)
	}()
	return path
}

func TestAControlRequestIsSentAsOneLineAndTheNodesLineIsReturned(t *testing.T) {
	t.Parallel()
	seen := make(chan []byte, 1)
	path := listenForControl(t, func(request []byte, conn net.Conn) {
		seen <- request
		_, _ = conn.Write([]byte(`{"result":"applied","state":{}}` + "\n"))
	})
	answer, err := SendControl(t.Context(), path, []byte(`{"op":"state"}`))
	if request := <-seen; err != nil || string(answer) != `{"result":"applied","state":{}}`+"\n" || string(request) != `{"op":"state"}`+"\n" {
		t.Fatalf("SendControl = %q, %v after the node read %q, want one line each way", answer, err, request)
	}
}

func TestANodeThatDoesNotAnswerInTimeIsSilentNotGone(t *testing.T) {
	t.Parallel()
	released := make(chan struct{})
	t.Cleanup(func() { close(released) })
	path := listenForControl(t, func([]byte, net.Conn) { <-released })
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if _, err := SendControl(ctx, path, []byte(`{"op":"state"}`)); !errors.Is(err, ErrControlSilent) {
		t.Fatalf("SendControl to a silent node = %v, want ErrControlSilent", err)
	}
}

func TestANodeThatHangsUpWithoutAnAnswerIsNotSilent(t *testing.T) {
	t.Parallel()
	path := listenForControl(t, func([]byte, net.Conn) {})
	if _, err := SendControl(t.Context(), path, []byte(`{"op":"state"}`)); err == nil || errors.Is(err, ErrControlSilent) {
		t.Fatalf("SendControl to a node that hung up = %v, want a read failure", err)
	}
}

func TestASocketThatIsNotThereIsAFailureToReach(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "absent.sock")
	if _, err := SendControl(t.Context(), missing, []byte(`{"op":"state"}`)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("SendControl to an absent socket = %v, want ErrNotExist", err)
	}
	if ControlReady(missing) {
		t.Fatal("an absent socket is not ready")
	}
}

func TestTheControlSocketLivesInAPrivateDirectoryThatIsRemovedWithIt(t *testing.T) {
	t.Parallel()
	path, remove, err := NewControlSocket()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Dir(path))
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 || filepath.Base(path) != controlSocket {
		t.Fatalf("control directory = %v, %v for %s, want an owner-only directory holding %s", info, err, path, controlSocket)
	}
	remove()
	if _, err := os.Lstat(filepath.Dir(path)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("after remove: %v, want the directory gone", err)
	}
}

func TestAControlSocketIsReadyOnlyWhileTheNodeListens(t *testing.T) {
	t.Parallel()
	path := listenForControl(t, func([]byte, net.Conn) {})
	if !ControlReady(path) {
		t.Fatal("a listening socket is ready")
	}
	if ControlReady(filepath.Dir(path)) {
		t.Fatal("a directory is not a socket")
	}
}
