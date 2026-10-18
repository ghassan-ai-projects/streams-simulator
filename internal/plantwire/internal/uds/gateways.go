package uds

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

type Gateways struct {
	listener *net.UnixListener
	path     string
	socket   os.FileInfo
	links    chan *Link
	done     chan struct{}
	wait     time.Duration
	stop     func() bool
}

func ListenGateways(ctx context.Context, path string, wait time.Duration) (*Gateways, error) {
	listener, err := kernel.ListenPrivate(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("plantwire relay: %w", err)
	}
	socket, _ := os.Lstat(path)
	g := &Gateways{listener: listener, path: path, socket: socket, links: make(chan *Link), done: make(chan struct{}), wait: wait}
	g.stop = context.AfterFunc(ctx, func() { _ = listener.Close() })
	go g.accept()
	return g, nil
}

func (g *Gateways) Links() <-chan *Link { return g.links }

func (g *Gateways) accept() {
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			return
		}
		if !g.handOver(NewAccepted(conn, g.wait)) {
			return
		}
	}
}

func (g *Gateways) handOver(link *Link) bool {
	select {
	case g.links <- link:
		return true
	case <-g.done:
		_ = link.Close()
		return false
	}
}

func (g *Gateways) Close() error {
	g.stop()
	select {
	case <-g.done:
	default:
		close(g.done)
	}
	err := g.listener.Close()
	g.removeOwnSocket()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("plantwire relay: close: %w", err)
	}
	return nil
}

func (g *Gateways) removeOwnSocket() {
	now, err := os.Lstat(g.path)
	if err == nil && now.Mode().Type() == os.ModeSocket && os.SameFile(now, g.socket) {
		_ = os.Remove(g.path)
	}
}

func Interrupted(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed)
}
