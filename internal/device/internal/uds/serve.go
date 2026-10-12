package uds

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
)

// Serve serves d on a unix socket, announces the banner, and returns after the
// process is interrupted and the listener is closed.
func Serve(stderr io.Writer, socket string, d *domain.Device, banner string) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)
	return serveUntil(stderr, socket, d, banner, stop)
}

// serveUntil serves d until anything arrives on stop.
func serveUntil(stderr io.Writer, socket string, d *domain.Device, banner string, stop <-chan os.Signal) error {
	listener, err := Listen(socket, d)
	if err != nil {
		return fmt.Errorf("streamsim: %w", listenFailure(socket, err))
	}
	_, _ = fmt.Fprintln(stderr, banner)
	return shutDown(stderr, listener, stop)
}

// listenFailure reports a socket that cannot be bound as a failure of its
// path, so the command names the socket and classifies the cause like any
// file-system failure.
func listenFailure(socket string, err error) error {
	var pathFailure *os.PathError
	if errors.As(err, &pathFailure) {
		return err
	}
	return &os.PathError{Op: "listen", Path: socket, Err: err}
}

func shutDown(stderr io.Writer, listener *net.UnixListener, stop <-chan os.Signal) error {
	<-stop
	_, _ = fmt.Fprintln(stderr, "streamsim device: shutting down")
	if err := listener.Close(); err != nil {
		return fmt.Errorf("streamsim: close device listener: %w", err)
	}
	return nil
}
