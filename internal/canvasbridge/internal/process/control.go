package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const (
	controlWait   = 12 * time.Second
	controlFolder = "streamsim-control-"
	controlSocket = "node.sock"
)

const maxSocketPath = 103

var (
	ErrControlSilent      = errors.New("canvasbridge: the node did not answer its control in time")
	ErrControlPathTooLong = errors.New("canvasbridge: no private directory is short enough for a unix socket path")
)

func NewControlSocket() (string, func(), error) {
	return makeControlSocket(os.TempDir(), "/tmp")
}

func makeControlSocket(bases ...string) (string, func(), error) {
	var failed error
	for _, base := range bases {
		path, remove, err := controlSocketIn(base)
		if err == nil {
			return path, remove, nil
		}
		failed = errors.Join(failed, err)
	}
	return "", nil, failed
}

func controlSocketIn(base string) (string, func(), error) {
	dir, err := os.MkdirTemp(base, controlFolder)
	if err != nil {
		return "", nil, fmt.Errorf("canvasbridge: a private directory for the control socket: %w", err)
	}
	path := dir + string(os.PathSeparator) + controlSocket
	if len(path) > maxSocketPath {
		_ = os.RemoveAll(dir)
		return "", nil, fmt.Errorf("%w: %d bytes under %s, at most %d", ErrControlPathTooLong, len(path), base, maxSocketPath)
	}
	return path, func() { _ = os.RemoveAll(dir) }, nil
}

func ControlReady(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().Type() == os.ModeSocket
}

func SendControl(ctx context.Context, path string, line []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, controlWait)
	defer cancel()
	conn, err := kernel.DialPrivate(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: dial the control socket: %w", err)
	}
	defer hangUpWhenDone(ctx, conn)()
	return exchange(ctx, conn, line)
}

func hangUpWhenDone(ctx context.Context, conn io.Closer) func() {
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	return func() {
		stop()
		_ = conn.Close()
	}
}

func exchange(ctx context.Context, conn io.ReadWriter, line []byte) ([]byte, error) {
	if _, err := conn.Write(append(line, '\n')); err != nil {
		return nil, fmt.Errorf("canvasbridge: write the control request: %w", err)
	}
	answer, err := bufio.NewReader(io.LimitReader(conn, domain.MaxControlReply)).ReadBytes('\n')
	switch {
	case err == nil:
		return answer, nil
	case ctx.Err() != nil:
		return nil, ErrControlSilent
	}
	return nil, fmt.Errorf("canvasbridge: read the control answer: %w", err)
}
