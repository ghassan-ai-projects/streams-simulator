package uds

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const (
	controlWait     = 10 * time.Second
	maxControlLine  = 4096
	maxControlConns = 8
)

type ControlHandler interface {
	Handle(ctx context.Context, line []byte) []byte
}

type ControlServer struct {
	listener *net.UnixListener
	path     string
	socket   os.FileInfo
	handler  ControlHandler
	slots    chan struct{}
	ctx      context.Context
	stop     func() bool
}

func ListenControl(ctx context.Context, path string, handler ControlHandler) (*ControlServer, error) {
	listener, err := kernel.ListenPrivate(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("plantwire control: %w", err)
	}
	socket, _ := os.Lstat(path)
	c := &ControlServer{listener: listener, path: path, socket: socket, handler: handler, slots: make(chan struct{}, maxControlConns), ctx: ctx}
	c.stop = context.AfterFunc(ctx, func() { _ = listener.Close() })
	go c.accept()
	return c, nil
}

func (c *ControlServer) accept() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			return
		}
		select {
		case c.slots <- struct{}{}:
			go c.serve(conn)
		default:
			_ = conn.Close()
		}
	}
}

func (c *ControlServer) serve(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		<-c.slots
	}()
	hangUp := context.AfterFunc(c.ctx, func() { _ = conn.Close() })
	defer hangUp()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, maxControlLine), maxControlLine)
	for scanner.Scan() {
		if !c.answer(conn, scanner.Bytes()) {
			return
		}
	}
}

func (c *ControlServer) answer(conn net.Conn, line []byte) bool {
	ctx, cancel := context.WithTimeout(c.ctx, controlWait)
	defer cancel()
	reply := c.handler.Handle(ctx, line)
	_, err := conn.Write(append(reply, '\n'))
	return err == nil
}

func (c *ControlServer) Close() error {
	c.stop()
	err := c.listener.Close()
	removeOwnSocket(c.path, c.socket)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("plantwire control: close: %w", err)
	}
	return nil
}
