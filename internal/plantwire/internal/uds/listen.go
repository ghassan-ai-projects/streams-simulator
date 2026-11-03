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
		return respond(ctx, link, peer, in, beats, delay)
	case <-beats:
		return beat(link, peer)
	}
}

func beat(link *Link, peer Peer) (heard, open bool, err error) {
	if err := answer(link, peer.Beat()); err != nil {
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

func respond(ctx context.Context, link *Link, peer Peer, in domain.Inbound, beats <-chan time.Time, delay time.Duration) (heard, open bool, err error) {
	if in.Type == "NSYNC" {
		if open, err := hold(ctx, link, peer, beats, delay); err != nil || !open {
			return true, open, err
		}
	}
	return true, true, reply(link, peer, in)
}

func hold(ctx context.Context, link *Link, peer Peer, beats <-chan time.Time, delay time.Duration) (open bool, err error) {
	if delay <= 0 {
		return true, nil
	}
	deadline := time.After(delay)
	for {
		over, open, err := heldTurn(ctx, link, peer, beats, deadline)
		if over || err != nil || !open {
			return open, err
		}
	}
}

func heldTurn(ctx context.Context, link *Link, peer Peer, beats, deadline <-chan time.Time) (over, open bool, err error) {
	select {
	case <-ctx.Done():
		return true, false, fmt.Errorf("plantwire host: %w", ctx.Err())
	case <-deadline:
		return true, true, nil
	case <-beats:
		_, open, err = beat(link, peer)
		return false, open, err
	}
}

func reply(link *Link, peer Peer, in domain.Inbound) error {
	replies := peer.Receive(in)
	if peer.Accepted() {
		if err := link.Settle(); err != nil {
			return err
		}
	}
	return answer(link, replies)
}

func answer(link *Link, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}
	return link.Send(records)
}
