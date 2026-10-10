// Package serve is the long-running-process edge of the cli module: the
// device socket that lives until interrupted and the MCP director session
// with its optional operator HTTP endpoint.
package serve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
)

// Device serves dev on a unix socket, announces the banner, and returns after
// the process is interrupted and the listener is closed.
func Device(stderr io.Writer, socket string, dev *device.Device, banner string) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)
	return deviceUntil(stderr, socket, dev, banner, stop)
}

// deviceUntil serves dev until anything arrives on stop.
func deviceUntil(stderr io.Writer, socket string, dev *device.Device, banner string, stop <-chan os.Signal) error {
	listener, err := device.Listen(socket, dev)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	_, _ = fmt.Fprintln(stderr, banner)
	return shutDown(stderr, listener, stop)
}

func shutDown(stderr io.Writer, listener *net.UnixListener, stop <-chan os.Signal) error {
	<-stop
	_, _ = fmt.Fprintln(stderr, "streamsim device: shutting down")
	if err := listener.Close(); err != nil {
		return fmt.Errorf("streamsim: close device listener: %w", err)
	}
	return nil
}

// Director serves the director role over stdio until the session ends. When
// operatorAddr is set the operator role is also bound over streamable HTTP
// and sim.world.create returns the endpoint with the token.
func Director(stderr io.Writer, d *mcp.Director, operatorAddr string) error {
	return directorOn(stderr, d, operatorAddr, &mcpsdk.StdioTransport{})
}

// directorOn serves the director role over transport.
func directorOn(stderr io.Writer, d *mcp.Director, operatorAddr string, transport mcpsdk.Transport) error {
	srv := mcp.NewDirectorServer(d)
	var httpServer *http.Server
	if operatorAddr != "" {
		var err error
		if httpServer, err = operatorEndpoint(stderr, d, operatorAddr); err != nil {
			return err
		}
	}
	return runSession(srv, httpServer, transport)
}

// operatorEndpoint binds the operator role over streamable HTTP so a consumer
// process can connect with the capability token from sim.world.create. The
// returned server is closed when the director stdio session ends.
func operatorEndpoint(stderr io.Writer, d *mcp.Director, addr string) (*http.Server, error) {
	opServer := mcp.NewOperatorServerResolver(d)
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return opServer }, nil)
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("mcp operator listener: %w", err)
	}
	endpoint := "http://" + ln.Addr().String()
	d.SetOperatorEndpoint(endpoint)
	_, _ = fmt.Fprintf(stderr, "operator endpoint: %s\n", endpoint)
	httpServer := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go serveOperator(stderr, httpServer, ln)
	return httpServer, nil
}

func serveOperator(stderr io.Writer, server *http.Server, listener net.Listener) {
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		_, _ = fmt.Fprintf(stderr, "operator endpoint failed: %v\n", err)
	}
}

func runSession(server *mcpsdk.Server, operator *http.Server, transport mcpsdk.Transport) error {
	err := server.Run(context.Background(), transport)
	if operator != nil {
		_ = operator.Close()
	}
	if err != nil {
		return fmt.Errorf("mcp director: %w", err)
	}
	return nil
}
