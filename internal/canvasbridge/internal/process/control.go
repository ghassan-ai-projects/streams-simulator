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

var ErrControlSilent = errors.New("canvasbridge: the node did not answer its control in time")

func NewControlSocket() (string, func(), error) {
	dir, err := os.MkdirTemp("", controlFolder)
	if err != nil {
		return "", nil, fmt.Errorf("canvasbridge: a private directory for the control socket: %w", err)
	}
	return dir + string(os.PathSeparator) + controlSocket, func() { _ = os.RemoveAll(dir) }, nil
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
