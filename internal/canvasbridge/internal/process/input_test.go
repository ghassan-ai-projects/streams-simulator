package process

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithInputDeliversTheBytesOnStandardInput(t *testing.T) {
	t.Parallel()
	cases := map[string][]byte{
		"text":        []byte(`{"kind":"fan"}`),
		"large input": []byte(strings.Repeat("0123456789", 100000)),
		"no bytes":    {},
		"no input":    nil,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result, err := RunWithInput(context.Background(), self(t), t.TempDir(), []string{"--helper", "cat"}, input)
			if err != nil || result.ExitCode != 0 || string(result.Stdout) != string(input) {
				t.Fatalf("RunWithInput = %d bytes out, exit %d, %v, want the %d bytes echoed", len(result.Stdout), result.ExitCode, err, len(input))
			}
		})
	}
}

func TestRunWithInputReportsACommandThatFailsAndAnExecutableThatIsAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	failed, err := RunWithInput(context.Background(), self(t), dir, []string{"--helper", "fail"}, []byte("ignored"))
	if err != nil || failed.ExitCode != 3 || !strings.HasPrefix(string(failed.Stderr), "refused") {
		t.Fatalf("a failing command: %+v, %v", failed, err)
	}
	if _, err := RunWithInput(context.Background(), filepath.Join(dir, "absent"), dir, nil, []byte("x")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing executable ran: %v", err)
	}
}

func TestRunWithInputStopsWhenItsContextIsCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := RunWithInput(ctx, self(t), t.TempDir(), []string{"--helper", "cat"}, []byte("x")); !errors.Is(err, context.Canceled) {
		t.Fatalf("RunWithInput with a canceled context = %v, want context.Canceled", err)
	}
}
