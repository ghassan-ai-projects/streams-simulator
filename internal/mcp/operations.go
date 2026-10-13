package mcp

import (
	"context"
	"io"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/protocol"
)

func ServeDirector(ctx context.Context, stderr io.Writer, options ServeOptions) error {
	return protocol.ServeDirector(ctx, stderr, options)
}

func NewOperatorServerResolver(d *director.Director) *mcpsdk.Server {
	return protocol.NewOperatorServerResolver(d)
}
