// Package process is the process edge of the canvas bridge: the director it
// drives over MCP (one `streamsim mcp --role director` child on stdio) and the
// streamsim commands it runs for CLI actions. Both run the bridge's own
// executable with a scrubbed environment; arguments are separate argv entries.
package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ErrGone = errors.New("canvasbridge: the director is gone")

type Director struct {
	session *mcpsdk.ClientSession
	stderr  *lastWords
}

type ToolResult struct {
	Structured json.RawMessage
	Failure    string
}

func StartDirector(ctx context.Context, executable string, args []string) (*Director, error) {
	command := directorCommand(ctx, executable, args)
	words, release, err := captureStderr(command)
	if err != nil {
		return nil, err
	}
	session, err := connect(ctx, command)
	release()
	if err != nil {
		return nil, fmt.Errorf("%w%s", err, words.said())
	}
	return &Director{session: session, stderr: words}, nil
}

func directorCommand(ctx context.Context, executable string, args []string) *exec.Cmd {
	//nolint:gosec // G204: the bridge's own executable with argv entries the bridge built (governance/DEFERRED.md D-64)
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = []string{}
	return command
}

func connect(ctx context.Context, command *exec.Cmd) (*mcpsdk.ClientSession, error) {
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "streamsim-canvas-bridge", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.CommandTransport{Command: command}, nil)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: start the director: %w", err)
	}
	return session, nil
}

func (d *Director) Call(ctx context.Context, tool string, args map[string]any) (ToolResult, error) {
	res, err := d.session.CallTool(ctx, &mcpsdk.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return ToolResult{}, d.callFailure(ctx, tool, err)
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

func (d *Director) callFailure(ctx context.Context, tool string, err error) error {
	if ctx.Err() != nil || !connectionLost(err) {
		return fmt.Errorf("canvasbridge: %s: %w", tool, err)
	}
	return fmt.Errorf("%w: %s: %w%s", ErrGone, tool, err, d.stderr.said())
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
