package transport

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestListenRefusesANonLoopbackAddress(t *testing.T) {
	t.Parallel()
	for _, addr := range []string{"0.0.0.0:0", ":0", "192.168.1.9:8080", "example.com:80", "[::]:0", "nonsense"} {
		listener, err := Listen(t.Context(), addr)
		if !errors.Is(err, domain.ErrNonLoopback) || listener != nil {
			t.Errorf("Listen(%q) = %v, %v; want ErrNonLoopback and no listener", addr, listener, err)
		}
	}
}

func TestListenBindsLoopbackAndReportsAnOccupiedPort(t *testing.T) {
	t.Parallel()
	listener, err := Listen(t.Context(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if ip := listener.Addr().(*net.TCPAddr).IP; !ip.IsLoopback() {
		t.Fatalf("bound %v", ip)
	}
	if _, err := Listen(t.Context(), listener.Addr().String()); !errors.Is(err, syscall.EADDRINUSE) {
		t.Fatalf("an occupied port: %v", err)
	}
}

func TestServeAnswersUntilItsContextEnds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	listener, err := Listen(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ui := t.TempDir()
	writeUI(t, ui, "index.html", "<html>canvas</html>")
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, listener, &fakeService{}, ui) }()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+listener.Addr().String()+"/api/session", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	cancel()
	if err := <-done; err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("Serve ended with %v after answering %d", err, response.StatusCode)
	}
}

func TestServeNeedsTheUIDirectory(t *testing.T) {
	t.Parallel()
	listener, err := Listen(t.Context(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if err := Serve(t.Context(), listener, &fakeService{}, t.TempDir()+"/absent"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing UI directory: %v", err)
	}
}
