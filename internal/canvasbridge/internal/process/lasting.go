package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type LastingExit struct {
	Stdout   []byte
	Stderr   string
	ExitCode int
}

type Lasting struct {
	command *exec.Cmd
	stdout  bytes.Buffer
	stderr  tail
	done    chan struct{}
	exit    LastingExit
}

func StartLasting(executable, dir string, argv []string, stdin []byte) (*Lasting, error) {
	command := directorCommand(context.Background(), executable, argv)
	lasting := &Lasting{command: command, done: make(chan struct{})}
	command.Dir, command.Stdin = dir, bytes.NewReader(stdin)
	command.Stdout, command.Stderr = &lasting.stdout, &lasting.stderr
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("canvasbridge: start streamsim %v: %w", argv, err)
	}
	go lasting.wait()
	return lasting, nil
}

func (l *Lasting) wait() {
	err := l.command.Wait()
	l.exit = LastingExit{Stdout: l.stdout.Bytes(), Stderr: l.stderr.String(), ExitCode: exitCodeOf(err)}
	close(l.done)
}

func exitCodeOf(err error) int {
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		return exit.ExitCode()
	case err != nil:
		return -1
	}
	return 0
}

func (l *Lasting) Exited() (LastingExit, bool) {
	select {
	case <-l.done:
		return l.exit, true
	default:
		return LastingExit{}, false
	}
}

func (l *Lasting) Stop(within time.Duration) LastingExit {
	_ = l.command.Process.Signal(os.Interrupt)
	timer := time.NewTimer(within)
	defer timer.Stop()
	select {
	case <-l.done:
	case <-timer.C:
		_ = l.command.Process.Kill()
		<-l.done
	}
	return l.exit
}
