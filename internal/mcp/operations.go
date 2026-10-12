package mcp

import (
	"context"
	"io"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/protocol"
)

// ServeDirector loads the director that options name and serves its role
// over stdio until the session ends. When options.OperatorAddr is set the
// operator role is also bound over streamable HTTP and sim.world.create
// returns the endpoint with the token.
func ServeDirector(ctx context.Context, stderr io.Writer, options ServeOptions) error {
	return protocol.ServeDirector(ctx, stderr, options)
}

// NewDirectorServer builds the MCP server of the director role: catalog,
// world, clock, fault, perturbation and truth tools.
func NewDirectorServer(d *director.Director) *mcpsdk.Server {
	return protocol.NewDirectorServer(d)
}

// NewOperatorServerResolver builds the MCP server of the operator role over
// a director: every call resolves its capability token to one world's
// operator view, so one endpoint serves every world.
func NewOperatorServerResolver(d *director.Director) *mcpsdk.Server {
	return protocol.NewOperatorServerResolver(d)
}
