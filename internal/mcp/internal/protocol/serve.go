package protocol

import (
	"context"
	"io"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/transport"
)

type ServeOptions struct {
	DomainsDir, AdaptersDir, OutDir, OperatorAddr string
	ComponentsDirs                                []string
}

func ServeDirector(ctx context.Context, stderr io.Writer, options ServeOptions) error {
	d, err := newDirector(ctx, options)
	if err != nil {
		return err
	}
	return serveDirectorOn(ctx, stderr, d, options.OperatorAddr, &mcpsdk.StdioTransport{})
}

func serveDirectorOn(ctx context.Context, stderr io.Writer, d *director.Director, operatorAddr string, t mcpsdk.Transport) error {
	srv := NewDirectorServer(d)
	endpoint, err := bindOperator(ctx, stderr, d, operatorAddr)
	if err != nil {
		return err
	}
	return transport.Serve(ctx, srv, endpoint, t)
}

func bindOperator(ctx context.Context, stderr io.Writer, d *director.Director, addr string) (*transport.Endpoint, error) {
	if addr == "" {
		return nil, nil
	}
	return transport.BindOperator(ctx, stderr, NewOperatorServerResolver(d), addr, d.SetOperatorEndpoint)
}

func newDirector(ctx context.Context, options ServeOptions) (*director.Director, error) {
	d, err := director.NewFromDirs(ctx, options.DomainsDir, options.AdaptersDir, options.OutDir)
	if err != nil || len(options.ComponentsDirs) == 0 {
		return d, err
	}
	if err := d.LoadLibrary(options.ComponentsDirs); err != nil {
		return nil, err
	}
	return d, nil
}
