package protocol

import (
	"context"
	"io"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/transport"
)

// ServeDirector loads the director options name and serves its role over stdio
// until the session ends. When options.OperatorAddr is set the operator role
// is also bound over streamable HTTP and sim.world.create returns the
// endpoint with the token.
func ServeDirector(ctx context.Context, stderr io.Writer, options app.ServeOptions) error {
	d, err := app.LoadDirector(ctx, options)
	if err != nil {
		return err
	}
	return serveDirectorOn(ctx, stderr, d, options.OperatorAddr, &mcpsdk.StdioTransport{})
}

// serveDirectorOn serves the director role of d over transport.
func serveDirectorOn(ctx context.Context, stderr io.Writer, d *app.Director, operatorAddr string, t mcpsdk.Transport) error {
	srv := NewDirectorServer(d)
	endpoint, err := bindOperator(ctx, stderr, d, operatorAddr)
	if err != nil {
		return err
	}
	return transport.Serve(ctx, srv, endpoint, t)
}

// bindOperator binds the operator role of d when an address is given.
func bindOperator(ctx context.Context, stderr io.Writer, d *app.Director, addr string) (*transport.Endpoint, error) {
	if addr == "" {
		return nil, nil
	}
	return transport.BindOperator(ctx, stderr, NewOperatorServerResolver(d), addr, d.SetOperatorEndpoint)
}
