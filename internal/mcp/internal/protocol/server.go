// Package protocol is the MCP driving adapter of the surface: it declares the
// tools and their closed input schemas, registers them with the SDK, decodes
// arguments and hands them to the app layer's use cases. It holds no
// simulator rule and is the only place the MCP SDK is imported.
package protocol

import (
	"context"
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var implementation = &mcp.Implementation{Name: "streamsim", Version: model.SimVersion}

type toolDef struct {
	name        string
	description string
	schema      json.RawMessage

	handler func(ctx context.Context, args map[string]any) (any, error)
}

func addTool(s *mcp.Server, def toolDef) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        def.name,
		Description: def.description,
		InputSchema: def.schema,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
		out, err := def.handler(ctx, args)
		return nil, out, err
	})
}
