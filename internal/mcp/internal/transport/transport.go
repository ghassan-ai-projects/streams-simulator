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

// Endpoint is the operator role bound over streamable HTTP. The zero value
// and nil are an endpoint that is not bound.
type Endpoint struct {
	server *http.Server
}

// BindOperator binds the operator role over streamable HTTP at addr so a
// consumer process can connect with the capability token from
// sim.world.create. Before serving it passes the endpoint URL to announce and
// prints it on stderr, and it returns the endpoint to close when the director
// session ends.
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

// Close stops the endpoint; closing an endpoint that is not bound does
// nothing.
func (e *Endpoint) Close() {
	if e != nil && e.server != nil {
		_ = e.server.Close()
	}
}

// Serve runs the director server over transport until the session ends, then
// closes the operator endpoint, if any.
func Serve(ctx context.Context, director *mcpsdk.Server, operator *Endpoint, transport mcpsdk.Transport) error {
	err := director.Run(ctx, transport)
	operator.Close()
	if err != nil {
		return fmt.Errorf("mcp director: %w", err)
	}
	return nil
}
