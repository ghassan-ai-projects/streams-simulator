// Package protocol is the MCP driving adapter of the surface: it declares the
// tools and their closed input schemas, registers them with the SDK, decodes
// arguments and hands them to the app layer's use cases. It holds no
// simulator rule and is the only place the MCP SDK is imported.
package protocol

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Implementation identity advertised by the server.
var implementation = &mcp.Implementation{Name: "streamsim", Version: "0.1.0"}

// toolDef describes one tool for the SDK wiring.
type toolDef struct {
	name        string
	description string
	schema      json.RawMessage
	// handler receives the decoded arguments and returns a JSON value.
	handler func(ctx context.Context, args map[string]any) (any, error)
}

// addTool registers one tool with the SDK. The SDK validates every call
// against the tool's declared schema before the handler runs; handlers keep
// the cross-field and domain-dependent checks the static schema cannot
// express (e.g. effector arguments validated against the loaded domain).
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
