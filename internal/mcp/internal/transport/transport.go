// Package transport is the serving edge of the mcp module: the listener that
// binds the operator role over streamable HTTP and the session that serves
// the director role until its transport closes.
package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Endpoint struct {
	server *http.Server
}

func BindOperator(ctx context.Context, stderr io.Writer, operator *mcpsdk.Server, addr string, announce func(endpoint string)) (*Endpoint, error) {
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return operator }, nil)
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("mcp operator listener: %w", err)
	}
	endpoint := "http://" + ln.Addr().String()
	announce(endpoint)
	_, _ = fmt.Fprintf(stderr, "operator endpoint: %s\n", endpoint)
	httpServer := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go serveOperator(stderr, httpServer, ln)
	return &Endpoint{server: httpServer}, nil
}

func serveOperator(stderr io.Writer, server *http.Server, listener net.Listener) {
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		_, _ = fmt.Fprintf(stderr, "operator endpoint failed: %v\n", err)
	}
}

func (e *Endpoint) Close() {
	if e != nil && e.server != nil {
		_ = e.server.Close()
	}
}

func Serve(ctx context.Context, director *mcpsdk.Server, operator *Endpoint, transport mcpsdk.Transport) error {
	err := director.Run(ctx, transport)
	operator.Close()
	if err != nil {
		return fmt.Errorf("mcp director: %w", err)
	}
	return nil
}
