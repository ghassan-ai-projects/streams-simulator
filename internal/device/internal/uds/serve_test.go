package uds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func dial(socket string) (net.Conn, error) {
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "unix", socket)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", socket, err)
	}
	return conn, nil
}

// readyWriter signals when the banner (written once the socket is bound)
// arrives.
type readyWriter struct {
	bytes.Buffer
	ready chan struct{}
}

func (w *readyWriter) Write(p []byte) (int, error) {
	n, _ := w.Buffer.Write(p)
	if bytes.Contains(p, []byte("listening")) {
		close(w.ready)
	}
	return n, nil
}

func TestDeviceListensUntilStoppedThenClosesItsSocket(t *testing.T) {
	t.Parallel()
	socket := testsupport.SocketPath(t)
	stderr := &readyWriter{ready: make(chan struct{})}
	stop := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() {
		done <- serveUntil(stderr, socket, domain.New(domain.Config{}), "streamsim device: listening on "+socket, stop)
	}()
	<-stderr.ready
	conn, err := dial(socket)
	if err != nil {
		t.Fatalf("the device must accept a link while served: %v", err)
	}
	_ = conn.Close()
	stop <- struct{}{}
	if err := <-done; err != nil {
		t.Fatalf("serve: %v", err)
	}
	if !strings.Contains(stderr.String(), "streamsim device: shutting down") {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if _, err := dial(socket); !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ECONNREFUSED) {
		t.Fatalf("dial after shutdown: err = %v, want the socket gone or refused", err)
	}
}

// A real interrupt: the handler is registered before the socket is bound, so
// an interrupt sent once the banner shows is always caught, never fatal.
//
//nolint:paralleltest // the interrupt is delivered to the whole test process
func TestDeviceStopsOnAnInterruptRegisteredBeforeBinding(t *testing.T) {
	socket := testsupport.SocketPath(t)
	stderr := &readyWriter{ready: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		done <- Serve(stderr, socket, domain.New(domain.Config{}), "streamsim device: listening on "+socket)
	}()
	<-stderr.ready
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("serve: %v", err)
	}
	if !strings.Contains(stderr.String(), "streamsim device: shutting down") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDeviceRefusesASocketPathItCannotBind(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	err := serveUntil(&stderr, blocker, domain.New(domain.Config{}), "banner", make(chan struct{}))
	if err == nil || !strings.HasPrefix(err.Error(), "streamsim: ") || stderr.Len() != 0 {
		t.Fatalf("err = %v, stderr = %q", err, stderr.String())
	}
}

func TestAListenFailureNamesTheSocketPath(t *testing.T) {
	t.Parallel()
	socket := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "absent", "d.sock")
	err := serveUntil(io.Discard, socket, domain.New(domain.Config{}), "banner", make(chan struct{}))
	var pathFailure *os.PathError
	if !errors.As(err, &pathFailure) || pathFailure.Path != socket || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("serveUntil = %v, want a path failure naming %s that is a missing directory", err, socket)
	}
	wrapped := listenFailure(socket, pathFailure)
	if !errors.Is(wrapped, pathFailure) {
		t.Fatal("an existing path failure must be returned as it is")
	}
}
