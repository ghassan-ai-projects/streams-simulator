package uds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const hangUpGrace = 200 * time.Millisecond

type Peer interface {
	Receive(domain.Inbound) []map[string]any
	Beat() []map[string]any
	Accepted() bool
	Done() bool
}

type Pacing struct {
	Delay     time.Duration
	Handshake time.Duration
	Heartbeat time.Duration
}

const HandshakeWait = 10 * time.Second

var ErrSocketInUse = kernel.ErrInUse

func ServeHost(ctx context.Context, path string, peer Peer, pacing Pacing, ready func()) error {
	ctx, stop := kernel.UntilInterrupted(ctx)
	defer stop()
	listener, err := kernel.ListenPrivate(ctx, path)
	if err != nil {
		return fmt.Errorf("plantwire host: %w", err)
	}
	defer func() { _, _ = listener.Close(), os.Remove(path) }()
	if ready != nil {
		ready()
	}
	return serveOne(ctx, listener, peer, pacing)
}

func serveOne(ctx context.Context, listener net.Listener, peer Peer, pacing Pacing) error {
	unblock := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer unblock()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("plantwire host: accept: %w", err)
		}
		if spoke, err := session(ctx, conn, peer, pacing); spoke || err != nil || peer.Done() {
			return err
		}
	}
}

func session(ctx context.Context, conn net.Conn, peer Peer, pacing Pacing) (bool, error) {
	link := NewAccepted(conn, HandshakeWait)
	defer func() { _ = link.Close() }()
	hangUp := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer hangUp()
	if err := link.Expect(pacing.Handshake); err != nil {
		return true, fmt.Errorf("plantwire host: %w", err)
	}
	beats, stop := heartbeats(pacing.Heartbeat)
	defer stop()
	return serveLink(ctx, link, peer, beats, pacing.Delay)
}

func heartbeats(every time.Duration) (<-chan time.Time, func()) {
	if every <= 0 {
		return nil, func() {}
	}
	return Ticks(every)
}

func serveLink(ctx context.Context, link *Link, peer Peer, beats <-chan time.Time, delay time.Duration) (bool, error) {
	spoke := false
	for !peer.Done() {
		heard, open, err := turn(ctx, link, peer, beats, delay)
		spoke = spoke || heard
		if err != nil || !open {
			return spoke, err
		}
	}
	return spoke, nil
}

func turn(ctx context.Context, link *Link, peer Peer, beats <-chan time.Time, delay time.Duration) (heard, open bool, err error) {
	select {
	case in, open := <-link.Incoming():
		if !open {
			return false, false, readEnd(link.failed)
		}
		return true, true, respond(ctx, link, peer, in, delay)
	case <-beats:
		return beat(ctx, link, peer)
	}
}

func beat(ctx context.Context, link *Link, peer Peer) (heard, open bool, err error) {
	if err := answer(ctx, link, peer.Beat(), 0); err != nil {
		heard, err = endOrFail(link, err)
		return heard, false, err
	}
	return false, true, nil
}

func endOrFail(link *Link, writeErr error) (heard bool, err error) {
	deadline := time.After(hangUpGrace)
	for {
		select {
		case _, open := <-link.Incoming():
			if !open {
				return heard, readEnd(link.failed)
			}
			heard = true
		case <-deadline:
			return heard, writeErr
		}
	}
}

func readEnd(failed error) error {
	if errors.Is(failed, io.EOF) {
		return nil
	}
	return fmt.Errorf("plantwire host: read: %w", failed)
}

func respond(ctx context.Context, link *Link, peer Peer, in domain.Inbound, delay time.Duration) error {
	if err := answer(ctx, link, peer.Receive(in), delay); err != nil {
		return err
	}
	if !peer.Accepted() {
		return nil
	}
	return link.Settle()
}

func answer(ctx context.Context, link *Link, records []map[string]any, delay time.Duration) error {
	for _, record := range records {
		if err := pause(ctx, record, delay); err != nil {
			return err
		}
	}
	if len(records) == 0 {
		return nil
	}
	return link.Send(records)
}

func pause(ctx context.Context, record map[string]any, delay time.Duration) error {
	if record["type"] != "HSYNC" || delay <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("plantwire host: %w", ctx.Err())
	case <-time.After(delay):
		return nil
	}
}
