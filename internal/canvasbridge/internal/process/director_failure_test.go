package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func addFailureTools(server *mcpsdk.Server, args []string) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.die_loudly"}, func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, map[string]any, error) {
		filler := strings.Repeat("x", 1023) + "\n"
		for range 128 {
			_, _ = os.Stderr.WriteString(filler)
		}
		_, _ = os.Stderr.WriteString("last words\n")
		syscall.Exit(9)
		return nil, nil, nil
	})
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.hang"}, func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, map[string]any, error) {
		writeOwnPID(args[0])
		_, _ = os.Stderr.WriteString("director hung\n")
		_ = os.Stdout.Close()
		<-time.After(time.Hour)
		return nil, nil, nil
	})
}

func writeOwnPID(path string) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }()
	_ = root.WriteFile(filepath.Base(path), []byte(strconv.Itoa(os.Getpid())), 0o600)
}

func TestADirectorThatDiesAlwaysReportsItsLastWords(t *testing.T) {
	t.Parallel()
	for batch := range 20 {
		t.Run(fmt.Sprintf("batch %d", batch), func(t *testing.T) {
			t.Parallel()
			for round := range 10 {
				director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director"})
				if err != nil {
					t.Fatal(err)
				}
				_, err = director.Call(context.Background(), "sim.die", nil)
				_ = director.Close()
				if !errors.Is(err, ErrGone) || !strings.Contains(err.Error(), "the director said: director crashed") {
					t.Fatalf("round %d: a director that exited during a call: %v", round, err)
				}
			}
		})
	}
}

func TestADirectorThatSaysMoreThanThePipeHoldsStillDiesWithItsLastWords(t *testing.T) {
	t.Parallel()
	director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = director.Close() }()
	_, err = director.Call(context.Background(), "sim.die_loudly", nil)
	if !errors.Is(err, ErrGone) || !strings.HasSuffix(err.Error(), "last words") {
		t.Fatalf("a director that wrote 128 KiB before exiting: %v", err)
	}
	if len(err.Error()) > tailLimit+256 {
		t.Fatalf("the report holds %d bytes, want only the last %d of what the director said", len(err.Error()), tailLimit)
	}
}

func TestADirectorThatNeverExitsIsReportedGoneAfterABoundedWaitAndStopped(t *testing.T) {
	t.Parallel()
	pidFile := t.TempDir() + "/pid"
	director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director", pidFile})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, err = director.Call(context.Background(), "sim.hang", nil)
	waited := time.Since(started)
	if !errors.Is(err, ErrGone) || !strings.HasSuffix(err.Error(), "the director said: director hung") {
		t.Fatalf("a director that closed its output and kept running: %v", err)
	}
	if waited < drainWait || waited > 10*drainWait {
		t.Fatalf("the report waited %s, want the bounded wait of %s", waited, drainWait)
	}
	_ = director.Close()
	assertProcessGone(t, pidFile)
}

func assertProcessGone(t *testing.T, pidFile string) {
	t.Helper()
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("process %d after Close: %v, want it stopped", pid, err)
	}
}

func TestEveryBrokenPipeErrorIsALostConnection(t *testing.T) {
	t.Parallel()
	cases := map[string]error{
		"connection closed": mcpsdk.ErrConnectionClosed,
		"end of file":       io.EOF,
		"closed pipe":       io.ErrClosedPipe,
		"file closed":       os.ErrClosed,
		"broken pipe":       errors.New("write |1: broken pipe"),
		"wrapped write":     fmt.Errorf("calling tools/call: %w", &os.PathError{Op: "write", Path: "|1", Err: errors.New("broken pipe")}),
		"wrapped closed":    fmt.Errorf("calling tools/call: %w", &os.PathError{Op: "write", Path: "|1", Err: os.ErrClosed}),
		"chain dropped":     errors.New("calling tools/call: " + (&os.PathError{Op: "write", Path: "|1", Err: errors.New("broken pipe")}).Error()),
		"closed file text":  errors.New("write |1: file already closed"),
		"closed network":    errors.New("use of closed network connection"),
	}
	for name, cause := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !connectionLost(cause) {
				t.Fatalf("connectionLost(%v) = false, want true", cause)
			}
		})
	}
	if connectionLost(errors.New("invalid params")) {
		t.Fatal("an unrelated error was taken for a lost connection")
	}
}

func TestConcurrentCallsOnADyingDirectorAllReportItsLastWords(t *testing.T) {
	t.Parallel()
	director, err := StartDirector(context.Background(), self(t), []string{"--helper", "director"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = director.Close() }()
	failures := make(chan error, 8)
	for range cap(failures) {
		go func() {
			_, err := director.Call(context.Background(), "sim.die", nil)
			failures <- err
		}()
	}
	for range cap(failures) {
		if err := <-failures; !errors.Is(err, ErrGone) || !strings.Contains(err.Error(), "director crashed") {
			t.Fatalf("a concurrent call on a dying director: %v", err)
		}
	}
}
