package process

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMain(m *testing.M) {
	if index := slices.Index(os.Args, "--helper"); index >= 0 {
		os.Exit(helper(os.Args[index+1:]))
	}
	os.Exit(m.Run())
}

func helper(args []string) int {
	switch args[0] {
	case "env":
		_, _ = os.Stdout.WriteString(strings.Join(os.Environ(), ","))
		return 0
	case "fail":
		_, _ = os.Stderr.WriteString("refused")
		return 3
	case "director":
		return serveHelperDirector()
	}
	return 2
}

func serveHelperDirector() int {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "helper", Version: "0"}, nil)
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.echo"}, func(_ context.Context, _ *mcpsdk.CallToolRequest, in map[string]any) (*mcpsdk.CallToolResult, map[string]any, error) {
		if in["refuse"] == true {
			return &mcpsdk.CallToolResult{IsError: true, Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: "world_not_found: no world"}}}, nil, nil
		}
		return nil, map[string]any{"echo": in["value"]}, nil
	})
	if err := server.Run(context.Background(), &mcpsdk.StdioTransport{}); err != nil {
		return 1
	}
	return 0
}

func self(t *testing.T) string {
	t.Helper()
	executable, err := Self()
	if err != nil {
		t.Fatal(err)
	}
	return executable
}

func TestCommandsRunWithAScrubbedEnvironmentInTheWorkspace(t *testing.T) {
	t.Setenv("STREAMSIM_SECRET", "leak")
	dir := t.TempDir()
	result, err := Run(context.Background(), self(t), dir, []string{"--helper", "env"})
	if err != nil || result.ExitCode != 0 || strings.Contains(string(result.Stdout), "STREAMSIM_SECRET") {
		t.Fatalf("Run = %+v, %v: the child sees no inherited variable", result, err)
	}
	failed, err := Run(context.Background(), self(t), dir, []string{"--helper", "fail"})
	if err != nil || failed.ExitCode != 3 || !strings.HasPrefix(string(failed.Stderr), "refused") {
		t.Fatalf("a failing command: %+v, %v", failed, err)
	}
	if _, err := Run(context.Background(), filepath.Join(dir, "absent"), dir, nil); err == nil {
		t.Fatal("a missing executable ran")
	}
}

func TestTheDirectorIsDrivenOverMCPOnStdio(t *testing.T) {
	t.Parallel()
	director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = director.Close() }()
	result, err := director.Call(context.Background(), "sim.echo", map[string]any{"value": "hi"})
	if err != nil || string(result.Structured) != `{"echo":"hi"}` {
		t.Fatalf("Call = %+v, %v", result, err)
	}
	refused, err := director.Call(context.Background(), "sim.echo", map[string]any{"refuse": true})
	if err != nil || refused.Failure != "world_not_found: no world" {
		t.Fatalf("a refused call: %+v, %v", refused, err)
	}
	if _, err := director.Call(context.Background(), "sim.absent", nil); err == nil {
		t.Fatal("an unknown tool was called")
	}
}

func TestTheTickerTicksUntilStoppedAndInterruptsEndTheContext(t *testing.T) {
	t.Parallel()
	ticks, stop := Ticker(time.Millisecond)
	<-ticks
	stop()
	ctx, cancel := UntilInterrupted(context.Background())
	cancel()
	<-ctx.Done()
}
