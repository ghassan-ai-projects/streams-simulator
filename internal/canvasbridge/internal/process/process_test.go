package process

import (
	"context"
	"errors"
	"io"
	"io/fs"
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
	case "cat":
		_, _ = io.Copy(os.Stdout, os.Stdin)
		return 0
	case "fail":
		_, _ = os.Stderr.WriteString("refused")
		return 3
	case "director":
		return serveHelperDirector(args[1:])
	case "stderr-exit":
		_, _ = os.Stderr.WriteString("domain factory.domain.json: invalid\n")
		return 3
	}
	return 2
}

func serveHelperDirector(args []string) int {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "helper", Version: "0"}, nil)
	addFailureTools(server, args)
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.die"}, func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, map[string]any, error) {
		_, _ = os.Stderr.WriteString("director crashed\n")
		os.Exit(9)
		return nil, nil, nil
	})
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
	t.Parallel()
	dir := t.TempDir()
	result, err := RunWithInput(context.Background(), self(t), dir, []string{"--helper", "env"}, nil)
	if err != nil || result.ExitCode != 0 || len(result.Stdout) != 0 {
		t.Fatalf("RunWithInput = %+v, %v: the child sees no inherited variable", result, err)
	}
	failed, err := RunWithInput(context.Background(), self(t), dir, []string{"--helper", "fail"}, nil)
	if err != nil || failed.ExitCode != 3 || !strings.HasPrefix(string(failed.Stderr), "refused") {
		t.Fatalf("a failing command: %+v, %v", failed, err)
	}
	if _, err := RunWithInput(context.Background(), filepath.Join(dir, "absent"), dir, nil, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing executable ran: %v", err)
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
	if _, err := director.Call(context.Background(), "sim.absent", nil); err == nil || !strings.Contains(err.Error(), "sim.absent") {
		t.Fatalf("an unknown tool was called: %v", err)
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

func TestADirectorThatDiesIsReportedGoneWithWhatItSaid(t *testing.T) {
	t.Parallel()
	director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = director.Close() }()
	if _, err := director.Call(context.Background(), "sim.die", nil); !errors.Is(err, ErrGone) || !strings.Contains(err.Error(), "director crashed") {
		t.Fatalf("a director that exited during a call: %v", err)
	}
	if _, err := director.Call(context.Background(), "sim.echo", nil); !errors.Is(err, ErrGone) {
		t.Fatalf("a call after the director exited: %v", err)
	}
}

func TestADirectorThatCannotStartSaysWhy(t *testing.T) {
	t.Parallel()
	_, err := StartDirector(context.Background(), self(t), []string{"--helper", "stderr-exit"})
	if err == nil || !strings.Contains(err.Error(), "factory.domain.json: invalid") {
		t.Fatalf("a director that exited on start: %v", err)
	}
}
