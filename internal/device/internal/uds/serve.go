package uds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

type interruptSource func(parent context.Context) (context.Context, context.CancelFunc)

func Serve(stderr io.Writer, socket string, d *domain.Device, banner string) error {
	return serveInterruptedBy(kernel.UntilInterrupted, stderr, socket, d, banner)
}

func serveInterruptedBy(source interruptSource, stderr io.Writer, socket string, d *domain.Device, banner string) error {
	ctx, stop := source(context.Background())
	defer stop()
	return serveUntil(stderr, socket, d, banner, ctx.Done())
}

func serveUntil(stderr io.Writer, socket string, d *domain.Device, banner string, stop <-chan struct{}) error {
	listener, err := Listen(socket, d)
	if err != nil {
		return fmt.Errorf("streamsim: %w", listenFailure(socket, err))
	}
	_, _ = fmt.Fprintln(stderr, banner)
	return shutDown(stderr, listener, stop)
}

func listenFailure(socket string, err error) error {
	var pathFailure *os.PathError
	if errors.As(err, &pathFailure) {
		return err
	}
	return &os.PathError{Op: "listen", Path: socket, Err: err}
}

func shutDown(stderr io.Writer, listener *net.UnixListener, stop <-chan struct{}) error {
	<-stop
	_, _ = fmt.Fprintln(stderr, "streamsim device: shutting down")
	if err := listener.Close(); err != nil {
		return fmt.Errorf("streamsim: close device listener: %w", err)
	}
	return nil
}
