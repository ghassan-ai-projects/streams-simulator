package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/protocol"
)

// NewDirectorServer builds the MCP server of the director role: catalog,
// world, clock, fault, perturbation and truth tools.
func NewDirectorServer(d *Director) *mcpsdk.Server {
	return protocol.NewDirectorServer(d.director)
}

// NewOperatorServerResolver builds the MCP server of the operator role over
// a director: every call resolves its capability token to one world's
// operator view, so one endpoint serves every world.
func NewOperatorServerResolver(d *Director) *mcpsdk.Server {
	return protocol.NewOperatorServerResolver(d.director)
}

// SetOperatorEndpoint records the operator HTTP endpoint this process serves;
// sim.world.create includes it in its response.
func (d *Director) SetOperatorEndpoint(endpoint string) {
	d.director.SetOperatorEndpoint(endpoint)
}
