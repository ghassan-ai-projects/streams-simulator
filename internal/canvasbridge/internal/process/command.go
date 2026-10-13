package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

func Run(ctx context.Context, executable, dir string, argv []string) (CommandResult, error) {
	command := exec.CommandContext(ctx, executable, argv...) //nolint:gosec // G204: the bridge's own executable with argv it built
	command.Dir, command.Env = dir, []string{}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	return exitOf(command.Run(), CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, argv)
}

func exitOf(err error, result CommandResult, argv []string) (CommandResult, error) {
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		result.ExitCode = exit.ExitCode()
	case err != nil:
		return result, fmt.Errorf("canvasbridge: run streamsim %v: %w", argv, err)
	}
	return result, nil
}
