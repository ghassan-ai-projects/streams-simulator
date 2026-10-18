// Package process is the operating-system edge of the cli module: the
// standard streams, the environment and the wall clock of the process a command
// runs in.
package process

import (
	"io"
	"os"
	"time"
)

type Env struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Now    func() time.Time
	Getenv func(string) string
}

func Standard() Env {
	return Env{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr, Now: time.Now, Getenv: os.Getenv}
}
