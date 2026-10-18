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
	g := &Gateways{listener: listener, path: path, links: make(chan *Link), done: make(chan struct{}), wait: wait}
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
		link := NewAccepted(conn, g.wait)
		select {
		case g.links <- link:
		case <-g.done:
			_ = link.Close()
			return
		}
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
	_ = os.Remove(g.path)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("plantwire relay: close: %w", err)
	}
	return nil
}
