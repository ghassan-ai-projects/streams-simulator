package app

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func emptyTrace(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func options(t *testing.T) Options {
	t.Helper()
	return Options{Trace: emptyTrace(t), Out: filepath.Join(t.TempDir(), "verdict.json"), Threshold: 4, Window: 30}
}

func TestConsumeWritesTheVerdictOfATraceWithoutAnOperatorEndpoint(t *testing.T) {
	t.Parallel()
	o := options(t)
	summary, err := Consume(o)
	if err != nil || summary.Out != o.Out || summary.Detections != 0 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	raw, err := os.ReadFile(o.Out)
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateVerdict(raw); err != nil {
		t.Fatalf("the verdict file must match its contract: %v", err)
	}
}

func TestConsumeNamesTheFileItCannotReadOrWrite(t *testing.T) {
	t.Parallel()
	missing := options(t)
	missing.Trace = filepath.Join(t.TempDir(), "absent.jsonl")
	if _, err := Consume(missing); !errors.Is(err, fs.ErrNotExist) || !strings.HasPrefix(err.Error(), "streamsim: ") {
		t.Fatalf("a missing trace: err = %v", err)
	}
	unwritable := options(t)
	unwritable.Out = filepath.Join(t.TempDir(), "absent", "verdict.json")
	var pathFailure *fs.PathError
	if _, err := Consume(unwritable); !errors.As(err, &pathFailure) || pathFailure.Path != unwritable.Out {
		t.Fatalf("an unwritable verdict: err = %v", err)
	}
}

func TestConsumeRefusesAnIncompleteOperatorEndpoint(t *testing.T) {
	t.Parallel()
	o := options(t)
	o.MCPEndpoint = "http://127.0.0.1:1"
	if _, err := Consume(o); err == nil || !strings.Contains(err.Error(), "required together") {
		t.Fatalf("an endpoint without token and run: err = %v", err)
	}
}

func TestConsumeReportsItsVerdictOverTheOperatorEndpoint(t *testing.T) {
	t.Parallel()
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "operator"}, nil)
	reported := make(chan struct{}, 1)
	tool := func(name string, result any) {
		mcpsdk.AddTool(server, &mcpsdk.Tool{Name: name, InputSchema: map[string]any{"type": "object"}},
			func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, any, error) {
				if name == "sim.consumer.report" {
					select {
					case reported <- struct{}{}:
					default:
					}
				}
				return nil, result, nil
			})
	}
	tool("sim.nameplate.read", map[string]any{"world_id": "w-1", "entities": []any{}, "channels": []any{}, "effectors": []any{}})
	tool("sim.consumer.report", map[string]any{"ok": true})
	endpoint := httptest.NewServer(mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return server }, nil))
	t.Cleanup(endpoint.Close)

	o := options(t)
	o.MCPEndpoint, o.Token, o.RunID = endpoint.URL, "tok", "r-1"
	if summary, err := Consume(o); err != nil || summary.Detections != 0 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	select {
	case <-reported:
	default:
		t.Fatal("the verdict must be reported through the operator endpoint")
	}
}
