// Package process is the process edge of the canvas bridge: the director it
// drives over MCP (one `streamsim mcp --role director` child on stdio) and the
// streamsim commands it runs for CLI actions. Both run the bridge's own
// executable with a scrubbed environment; arguments are separate argv entries.
package process

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Director struct {
	session *mcpsdk.ClientSession
}

type ToolResult struct {
	Structured json.RawMessage
	Failure    string
}

func StartDirector(ctx context.Context, executable string, args []string) (*Director, error) {
	command := exec.Command(executable, args...) //nolint:gosec // G204: the bridge's own executable with argv it built
	command.Env = []string{}
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "streamsim-canvas-bridge", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.CommandTransport{Command: command}, nil)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: start the director: %w", err)
	}
	return &Director{session: session}, nil
}

func (d *Director) Call(ctx context.Context, tool string, args map[string]any) (ToolResult, error) {
	res, err := d.session.CallTool(ctx, &mcpsdk.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return ToolResult{}, fmt.Errorf("canvasbridge: %s: %w", tool, err)
	}
	if res.IsError {
		return ToolResult{Failure: textOf(res)}, nil
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		return ToolResult{}, fmt.Errorf("canvasbridge: %s result: %w", tool, err)
	}
	return ToolResult{Structured: raw}, nil
}

func (d *Director) Close() error {
	if err := d.session.Close(); err != nil {
		return fmt.Errorf("canvasbridge: stop the director: %w", err)
	}
	return nil
}

func textOf(res *mcpsdk.CallToolResult) string {
	for _, content := range res.Content {
		if text, ok := content.(*mcpsdk.TextContent); ok && text.Text != "" {
			return text.Text
		}
	}
	return "the director refused the call"
}
