package kernel

import (
	"context"
	"io"
	"net"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/process"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/sockets"
)

func OpenIfPresent(path string) (io.ReadCloser, bool, error) {
	return files.OpenIfPresent(path)
}

func Read(path string) ([]byte, error) {
	return files.Read(path)
}

func ReadRegular(path string, limit int64) ([]byte, error) {
	return files.ReadRegular(path, limit)
}

func ReadIfPresent(path string) ([]byte, bool, error) {
	return files.ReadIfPresent(path)
}

func Names(dir string) ([]string, error) {
	return files.Names(dir)
}

func PrivateDir(dir string) error {
	return files.PrivateDir(dir)
}

func WriteAtomic(path string, data []byte) error {
	return files.WriteAtomic(path, data)
}

func Replace(path string, data []byte) error {
	return files.Replace(path, data)
}

func Listen(ctx context.Context, path string) (*net.UnixListener, error) {
	return sockets.Listen(ctx, path)
}

func ListenPrivate(ctx context.Context, path string) (*net.UnixListener, error) {
	return sockets.ListenPrivate(ctx, path)
}

func DialPrivate(ctx context.Context, path string) (net.Conn, error) {
	return sockets.DialPrivate(ctx, path)
}

func UntilInterrupted(ctx context.Context) (context.Context, context.CancelFunc) {
	return process.UntilInterrupted(ctx)
}
